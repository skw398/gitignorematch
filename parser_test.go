package gitignorematch

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
)

func TestParseReader(t *testing.T) {
	m, err := ParseReader(strings.NewReader("*.tmp\n!keep.tmp\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("keep.tmp", false); got != Include {
		t.Fatalf("Match() = %s, want Include", got)
	}
}

func TestParseWithOptions(t *testing.T) {
	m, err := ParseWithOptions("Build/\n[A-Z].log\n[[:upper:]].txt\n[[:UPPER:]].invalid\n!KEEP.log\n", Options{CaseInsensitive: true})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path  string
		isDir bool
		want  Decision
	}{
		{path: "build/output", want: Ignore},
		{path: "BUILD/Output", want: Ignore},
		{path: "nested/bUiLd/OUTPUT", want: Ignore},
		{path: "a.log", want: Ignore},
		{path: "A.LOG", want: Ignore},
		{path: "nested/A.LoG", want: Ignore},
		{path: "a.txt", want: Ignore},
		{path: "A.TXT", want: Ignore},
		{path: "a.invalid", want: NoMatch},
		{path: "keep.log", want: Include},
		{path: "KeEp.LoG", want: Include},
		{path: "\u00e9.txt", want: NoMatch},
	}
	for _, tt := range tests {
		if got := m.Match(tt.path, tt.isDir); got != tt.want {
			t.Errorf("Match(%q, %t) = %s, want %s", tt.path, tt.isDir, got, tt.want)
		}
	}

	defaultMatcher, err := Parse("Build/")
	if err != nil {
		t.Fatal(err)
	}
	if got := defaultMatcher.Match("build/output", false); got != NoMatch {
		t.Fatalf("default Match() = %s, want NoMatch", got)
	}
}

func TestOptionedParsers(t *testing.T) {
	options := Options{CaseInsensitive: true}
	fromReader, err := ParseReaderWithOptions(strings.NewReader("Build/\n"), options)
	if err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(file, []byte("Build/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fromFile, err := ParseFileWithOptions(file, options)
	if err != nil {
		t.Fatal(err)
	}

	for _, matcher := range []*Matcher{fromReader, fromFile} {
		if got := matcher.Match("build/output", false); got != Ignore {
			t.Fatalf("Match() = %s, want Ignore", got)
		}
	}
}

func TestParseFile(t *testing.T) {
	const rules = "*.tmp\n!keep.tmp\n"
	file := filepath.Join(t.TempDir(), ".gitignore")
	if err := os.WriteFile(file, []byte(rules), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := ParseFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("other.tmp", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore", got)
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := ParseReader(nil); err == nil || !strings.Contains(err.Error(), "nil reader") {
		t.Fatalf("ParseReader(nil) error = %v, want nil reader error", err)
	}

	readErr := errors.New("reader failed")
	if _, err := ParseReader(iotest.ErrReader(readErr)); !errors.Is(err, readErr) {
		t.Fatalf("ParseReader(error reader) error = %v, want wrapped %v", err, readErr)
	}

	if _, err := ParseFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("ParseFile(missing) returned nil error")
	}
}

func TestEmptyPatternsAreIgnored(t *testing.T) {
	m, err := Parse("   \n!\n/\n# comment\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("anything", false); got != NoMatch {
		t.Fatalf("Match() = %s, want NoMatch", got)
	}
}

func TestCommentsBlankLinesAndCRLF(t *testing.T) {
	m, err := Parse("# comment\n\n*.tmp\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("x.tmp", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore", got)
	}
}
