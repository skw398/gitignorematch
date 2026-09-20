package gitignorematch

import "testing"

func TestRegressionNULTerminatesPatternLine(t *testing.T) {
	m, err := Parse("ignored\x00not-a-rule\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("ignored", false); got != Ignore {
		t.Fatalf("Match(ignored) = %s, want Ignore", got)
	}
	if got := m.Match("not-a-rule", false); got != NoMatch {
		t.Fatalf("Match(not-a-rule) = %s, want NoMatch", got)
	}
}

func TestRegressionIgnoredParentBlocksNegation(t *testing.T) {
	m, err := Parse("cache\n!cache/keep.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("cache/keep.txt", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore", got)
	}
}

func TestRegressionReopenedParentRestoresEarlierChildNegation(t *testing.T) {
	m, err := Parse("cache/\n!cache/keep.txt\n!cache/")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("cache/keep.txt", false); got != Include {
		t.Fatalf("Match() = %s, want Include", got)
	}
}

func TestRegressionNegatedDirectoryDoesNotIncludeDescendants(t *testing.T) {
	m, err := Parse("cache\n!cache/")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("cache/keep.txt", false); got != NoMatch {
		t.Fatalf("Match() = %s, want NoMatch", got)
	}
}

func TestRegressionGitCharacterClassEdges(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    Decision
	}{
		{name: "literal closing bracket", pattern: "[]]", path: "]", want: Ignore},
		{name: "POSIX class", pattern: "[[:digit:]]", path: "7", want: Ignore},
		{name: "caret negation", pattern: "[^a]", path: "b", want: Ignore},
		{name: "unterminated class", pattern: "foo[", path: "foo[", want: NoMatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Parse(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if got := m.Match(tt.path, false); got != tt.want {
				t.Fatalf("Match() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRegressionUnknownPOSIXClassNeverMatches(t *testing.T) {
	m, err := Parse("[[:spaci:]]")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("s", false); got != NoMatch {
		t.Fatalf("Match() = %s, want NoMatch", got)
	}
}

func TestRegressionDescendingRangeMatchesOnlyStart(t *testing.T) {
	m, err := Parse("[z-a]")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("z", false); got != Ignore {
		t.Fatalf("range start Match() = %s, want Ignore", got)
	}
	if got := m.Match("a", false); got != NoMatch {
		t.Fatalf("range end Match() = %s, want NoMatch", got)
	}
}

func TestRegressionCharacterClassDoesNotMatchSlash(t *testing.T) {
	m, err := Parse("foo[/]bar")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("foo/bar", false); got != NoMatch {
		t.Fatalf("Match() = %s, want NoMatch", got)
	}
}

func TestRegressionEscapedSeparatorAfterGlobstar(t *testing.T) {
	m, err := Parse(`foo/**\/bar`)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("foo/x/y/bar", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore", got)
	}
	if got := m.Match("foo/bar", false); got != NoMatch {
		t.Fatalf("zero-directory Match() = %s, want NoMatch", got)
	}
}

func TestRegressionNonUTF8Pattern(t *testing.T) {
	m, err := Parse("\xff*")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("\xffname", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore", got)
	}
}

func TestRegressionEscapedSlashRemainsRootRelative(t *testing.T) {
	m, err := Parse(`foo\/bar`)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("foo/bar", false); got != Ignore {
		t.Fatalf("root path Match() = %s, want Ignore", got)
	}
	if got := m.Match("nested/foo/bar", false); got != NoMatch {
		t.Fatalf("nested path Match() = %s, want NoMatch", got)
	}
}

func TestRegressionQuestionWildcardMatchesOneByte(t *testing.T) {
	m, err := Parse("日本?.txt\n日本???.log")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("日本語.txt", false); got != NoMatch {
		t.Fatalf("single question Match() = %s, want NoMatch", got)
	}
	if got := m.Match("日本語.log", false); got != Ignore {
		t.Fatalf("three questions Match() = %s, want Ignore", got)
	}
}

func TestRegressionUTF8BOM(t *testing.T) {
	m, err := Parse("\uFEFF*.tmp")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("file.tmp", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore", got)
	}
}

func TestRegressionDanglingEscapeNeverMatches(t *testing.T) {
	m, err := Parse("dangling\\\n*.valid")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("dangling\\", false); got != NoMatch {
		t.Fatalf("dangling escape Match() = %s, want NoMatch", got)
	}
	if got := m.Match("file.valid", false); got != Ignore {
		t.Fatalf("valid pattern Match() = %s, want Ignore", got)
	}
}
