package gitignorematch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const compatibilityFixturePath = "testdata/gitignore-compat.json"

type compatibilitySuite struct {
	PathScope  string                 `json:"path_scope"`
	CaseFormat []string               `json:"case_format"`
	Fixtures   []compatibilityFixture `json:"fixtures"`
}

type compatibilityFixture struct {
	Name            string              `json:"name"`
	Patterns        string              `json:"patterns"`
	CaseInsensitive bool                `json:"case_insensitive,omitempty"`
	Cases           []compatibilityCase `json:"cases"`
}

type compatibilityCase struct {
	Path  string
	IsDir bool
	Want  Decision
}

type gitCheckIgnoreResult struct {
	Decision Decision
	Pattern  string
	Line     int
}

func (c *compatibilityCase) UnmarshalJSON(data []byte) error {
	var fields []json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if len(fields) != 3 {
		return fmt.Errorf("want [path, is_dir, decision], got %d fields", len(fields))
	}
	if err := json.Unmarshal(fields[0], &c.Path); err != nil {
		return fmt.Errorf("path: %w", err)
	}
	if err := json.Unmarshal(fields[1], &c.IsDir); err != nil {
		return fmt.Errorf("is_dir: %w", err)
	}
	var want string
	if err := json.Unmarshal(fields[2], &want); err != nil {
		return fmt.Errorf("decision: %w", err)
	}
	switch want {
	case "ignore":
		c.Want = Ignore
	case "include":
		c.Want = Include
	case "no_match":
		c.Want = NoMatch
	default:
		return fmt.Errorf("unknown decision %q", want)
	}
	return nil
}

func TestGitCheckIgnoreCompatibility(t *testing.T) {
	suite := loadCompatibilitySuite(t)
	if suite.PathScope != "matcher-relative" {
		t.Fatalf("fixture path scope = %q, want matcher-relative", suite.PathScope)
	}
	if got := strings.Join(suite.CaseFormat, ","); got != "path,is_dir,decision" {
		t.Fatalf("fixture case format = %q, want path,is_dir,decision", got)
	}

	gitAvailable := true
	if _, err := exec.LookPath("git"); err != nil {
		gitAvailable = false
	}

	for _, fixture := range suite.Fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			m, err := ParseWithOptions(fixture.Patterns, Options{CaseInsensitive: fixture.CaseInsensitive})
			if err != nil {
				t.Fatalf("ParseWithOptions() error = %v", err)
			}
			for _, testCase := range fixture.Cases {
				t.Run(testCase.Path, func(t *testing.T) {
					if got := m.Match(testCase.Path, testCase.IsDir); got != testCase.Want {
						t.Fatalf("Match(%q, %t) = %s, want %s", testCase.Path, testCase.IsDir, got, testCase.Want)
					}
				})
			}

			if !gitAvailable {
				return
			}
			got := gitCheckIgnoreResults(t, fixture.Patterns, fixture.Cases, fixture.CaseInsensitive)
			for i, testCase := range fixture.Cases {
				if got[i].Decision != testCase.Want {
					t.Fatalf("git check-ignore(%q, %t) = %s, fixture expects %s", testCase.Path, testCase.IsDir, got[i].Decision, testCase.Want)
				}
				explanation := m.Explain(testCase.Path, testCase.IsDir)
				if explanation.Decision != got[i].Decision || explanation.Pattern != got[i].Pattern || explanation.Line != got[i].Line || explanation.Negated != strings.HasPrefix(got[i].Pattern, "!") {
					t.Fatalf("Explain(%q, %t) = %#v, git check-ignore = %#v", testCase.Path, testCase.IsDir, explanation, got[i])
				}
			}
		})
	}
}

func loadCompatibilitySuite(t *testing.T) compatibilitySuite {
	t.Helper()
	data, err := os.ReadFile(compatibilityFixturePath)
	if err != nil {
		t.Fatal(err)
	}
	var suite compatibilitySuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("decode %s: %v", compatibilityFixturePath, err)
	}
	return suite
}

func gitCheckIgnoreResults(
	t *testing.T,
	patterns string,
	cases []compatibilityCase,
	caseInsensitive bool,
) []gitCheckIgnoreResult {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(patterns), 0o644); err != nil {
		t.Fatal(err)
	}
	initRepo := exec.Command("git", "init", "-q")
	initRepo.Dir = dir
	if output, err := initRepo.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, strings.TrimSpace(string(output)))
	}

	var stdin bytes.Buffer
	for _, testCase := range cases {
		fullPath := filepath.Join(dir, filepath.FromSlash(testCase.Path))
		if testCase.IsDir {
			if err := os.MkdirAll(fullPath, 0o755); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(fullPath, nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		stdin.WriteString(testCase.Path)
		stdin.WriteByte(0)
	}

	ignoreCase := "core.ignoreCase=false"
	if caseInsensitive {
		ignoreCase = "core.ignoreCase=true"
	}
	cmd := exec.Command(
		"git", "-c", ignoreCase, "check-ignore",
		"--no-index", "--verbose", "--non-matching", "--stdin", "-z",
	)
	cmd.Dir = dir
	cmd.Stdin = &stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			t.Fatalf("git check-ignore: %v: %s", err, strings.TrimSpace(stderr.String()))
		}
	}

	fields := bytes.Split(stdout.Bytes(), []byte{0})
	if len(fields) != len(cases)*4+1 || len(fields[len(fields)-1]) != 0 {
		t.Fatalf("git check-ignore returned malformed output %q", stdout.Bytes())
	}
	results := make([]gitCheckIgnoreResult, len(cases))
	for i := range cases {
		pattern := string(fields[i*4+2])
		switch {
		case pattern == "":
			results[i].Decision = NoMatch
		case strings.HasPrefix(pattern, "!"):
			results[i].Decision = Include
		default:
			results[i].Decision = Ignore
		}
		if pattern == "" {
			continue
		}
		line, err := strconv.Atoi(string(fields[i*4+1]))
		if err != nil {
			t.Fatalf("git check-ignore returned invalid line number %q: %v", fields[i*4+1], err)
		}
		results[i].Pattern = pattern
		results[i].Line = line
	}
	return results
}
