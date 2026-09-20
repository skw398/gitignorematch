package gitignorematch

import "testing"

func FuzzMatch(f *testing.F) {
	f.Add("*.log", "app.log", false, false)
	f.Add("a/**/b", "a/x/b", false, false)
	f.Add("build/\n!build/keep.txt", "build/keep.txt", false, false)
	f.Add("[!a-z]??", "123", false, false)
	f.Add("[[:digit:]]", "7", false, false)
	f.Add("foo[", "foo[", false, false)
	f.Add("日本?.txt", "日本語.txt", false, false)
	f.Add("\uFEFF*.tmp", "file.tmp", false, false)
	f.Add("dangling\\", "dangling\\", false, false)
	f.Add("ignored\x00not-a-rule", "ignored", false, false)
	f.Add("Build/", "build/output", false, true)
	f.Add("\xff", "file", false, false)

	f.Fuzz(func(t *testing.T, patterns, path string, isDir, caseInsensitive bool) {
		m, err := ParseWithOptions(patterns, Options{CaseInsensitive: caseInsensitive})
		if err != nil {
			return
		}
		first := m.Match(path, isDir)
		if first != NoMatch && first != Ignore && first != Include {
			t.Fatalf("invalid decision: %d", first)
		}
		second := m.Match(path, isDir)
		if first != second {
			t.Fatalf("non-deterministic result: first=%s second=%s", first, second)
		}
	})
}
