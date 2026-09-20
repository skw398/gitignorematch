package gitignorematch

import (
	"math/rand"
	"os/exec"
	"strings"
	"testing"
)

func TestGeneratedGitCompatibility(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	rng := rand.New(rand.NewSource(1))
	for batch := 0; batch < 64; batch++ {
		patterns := generatedPatterns(rng)
		cases := generatedCases(rng)
		matcher, err := Parse(patterns)
		if err != nil {
			t.Fatalf("batch %d: Parse(%q): %v", batch, patterns, err)
		}
		want := gitCheckIgnoreResults(t, patterns, cases, false)
		for i, testCase := range cases {
			if got := matcher.Match(testCase.Path, testCase.IsDir); got != want[i].Decision {
				t.Fatalf("batch %d: patterns %q, Match(%q, %t) = %s, git check-ignore = %s", batch, patterns, testCase.Path, testCase.IsDir, got, want[i].Decision)
			}
		}
	}
}

func generatedPatterns(rng *rand.Rand) string {
	segments := []string{"a", "b", "build", "cache", "foo", "keep", "tmp"}
	extensions := []string{"log", "tmp", "txt"}

	count := 1 + rng.Intn(5)
	rules := make([]string, count)
	for i := range rules {
		segment := segments[rng.Intn(len(segments))]
		other := segments[rng.Intn(len(segments))]
		extension := extensions[rng.Intn(len(extensions))]
		var rule string
		switch rng.Intn(9) {
		case 0:
			rule = "*." + extension
		case 1:
			rule = segment + "/"
		case 2:
			rule = segment + "/*"
		case 3:
			rule = segment + "/**/" + other
		case 4:
			rule = "**/" + segment
		case 5:
			rule = segment + "/**"
		case 6:
			rule = "[ab]" + segment
		case 7:
			rule = segment + "?." + extension
		default:
			rule = "\\#" + segment
		}
		if rng.Intn(4) == 0 {
			rule = "!" + rule
		}
		rules[i] = rule
	}
	return strings.Join(rules, "\n") + "\n"
}

func generatedCases(rng *rand.Rand) []compatibilityCase {
	paths := []compatibilityCase{
		{Path: "app.log"},
		{Path: "keep.log"},
		{Path: "cache", IsDir: true},
		{Path: "cache/file"},
		{Path: "a/cache/file"},
		{Path: "build/tmp"},
		{Path: "foo/bar"},
		{Path: "foo/x/bar"},
		{Path: "afoo"},
		{Path: "bbuild"},
		{Path: "a.txt"},
		{Path: "b.tmp"},
		{Path: "#cache"},
		{Path: "notes.md"},
	}
	rng.Shuffle(len(paths), func(i, j int) {
		paths[i], paths[j] = paths[j], paths[i]
	})
	return paths[:6+rng.Intn(len(paths)-5)]
}
