package gitignorematch

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		patterns string
		path     string
		isDir    bool
		want     Decision
	}{
		{name: "wildcard file", patterns: "*.log", path: "app.log", want: Ignore},
		{name: "basename wildcard matches nested file", patterns: "*.log", path: "logs/app.log", want: Ignore},
		{name: "directory basename at any depth", patterns: "build/", path: "src/build", isDir: true, want: Ignore},
		{name: "directory rule applies to contents", patterns: "build/", path: "build/app.go", want: Ignore},
		{name: "directory suffix does not match file itself", patterns: "build/", path: "build", want: NoMatch},
		{name: "root anchored", patterns: "/secret.txt", path: "secret.txt", want: Ignore},
		{name: "root anchored not nested", patterns: "/secret.txt", path: "tmp/secret.txt", want: NoMatch},
		{name: "slash pattern is root relative", patterns: "doc/frotz", path: "doc/frotz", want: Ignore},
		{name: "slash pattern does not float", patterns: "doc/frotz", path: "a/doc/frotz", want: NoMatch},
		{name: "double star leading", patterns: "**/cache", path: "a/b/cache", isDir: true, want: Ignore},
		{name: "double star zero directories", patterns: "a/**/b", path: "a/b", want: Ignore},
		{name: "double star directories", patterns: "a/**/b", path: "a/x/y/b", want: Ignore},
		{name: "trailing double star", patterns: "dist/**", path: "dist", isDir: true, want: NoMatch},
		{name: "trailing double star contents", patterns: "dist/**", path: "dist/a/b.js", want: Ignore},
		{name: "negation", patterns: "*.log\n!important.log", path: "important.log", want: Include},
		{name: "escaped comment", patterns: `\#literal`, path: "#literal", want: Ignore},
		{name: "escaped bang", patterns: `\!literal`, path: "!literal", want: Ignore},
		{name: "trailing escaped space", patterns: "name\\ ", path: "name ", want: Ignore},
		{name: "trailing unescaped spaces", patterns: "name   ", path: "name", want: Ignore},
		{name: "question mark", patterns: "file?.txt", path: "file1.txt", want: Ignore},
		{name: "character class", patterns: "file[0-9].txt", path: "file7.txt", want: Ignore},
		{name: "negated character class", patterns: "file[!0-9].txt", path: "filex.txt", want: Ignore},
		{name: "dash at class edge", patterns: "file[a-].txt", path: "file-.txt", want: Ignore},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := Parse(tt.patterns)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got := m.Match(tt.path, tt.isDir); got != tt.want {
				t.Fatalf("Match(%q, %t) = %s, want %s", tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestNilMatcher(t *testing.T) {
	var nilMatcher *Matcher
	if got := nilMatcher.Match("file.tmp", false); got != NoMatch {
		t.Fatalf("nil matcher returned %s, want NoMatch", got)
	}
}

func TestExplain(t *testing.T) {
	m, err := Parse("# comment\n*.log\n!important.log\nbuild/\n!build/keep.log\n")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		path  string
		isDir bool
		want  MatchResult
	}{
		{
			name: "negated rule",
			path: "important.log",
			want: MatchResult{Decision: Include, Pattern: "!important.log", Line: 3, Negated: true},
		},
		{
			name: "ignored parent directory",
			path: "build/keep.log",
			want: MatchResult{Decision: Ignore, Pattern: "build/", Line: 4},
		},
		{
			name: "no matching rule",
			path: "notes.txt",
			want: MatchResult{Decision: NoMatch},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Explain(tt.path, tt.isDir); got != tt.want {
				t.Fatalf("Explain(%q, %t) = %#v, want %#v", tt.path, tt.isDir, got, tt.want)
			}
			if got := m.Match(tt.path, tt.isDir); got != tt.want.Decision {
				t.Fatalf("Match(%q, %t) = %s, want %s", tt.path, tt.isDir, got, tt.want.Decision)
			}
		})
	}
}

func TestPathNormalization(t *testing.T) {
	m, err := Parse("*.tmp")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", ".", "/"} {
		if got := m.Match(path, false); got != NoMatch {
			t.Errorf("Match(%q) = %s, want NoMatch", path, got)
		}
	}
	for _, path := range []string{"./file.tmp", "a/../file.tmp"} {
		if got := m.Match(path, false); got != Ignore {
			t.Errorf("Match(%q) = %s, want Ignore", path, got)
		}
	}
	for _, path := range []string{"/file.tmp", "../file.tmp", "a/../../file.tmp"} {
		if got := m.Match(path, false); got != NoMatch {
			t.Errorf("Match(%q) = %s, want NoMatch", path, got)
		}
	}
}

func TestLastMatchingRuleWins(t *testing.T) {
	m, err := Parse("*.log\n!important.log\nimportant.log")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Match("important.log", false); got != Ignore {
		t.Fatalf("Match() = %s, want Ignore from final matching rule", got)
	}
}
