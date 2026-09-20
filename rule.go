package gitignorematch

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type rule struct {
	pattern       string
	line          int
	negated       bool
	directoryOnly bool
	matchBasename bool
	regex         *regexp.Regexp
	literalPrefix string
}

func parseRule(line string, caseInsensitive bool) (rule, bool, error) {
	if line == "" || line[0] == '#' {
		return rule{}, false, nil
	}

	// Unescaped trailing spaces are insignificant in a .gitignore pattern.
	for len(line) > 0 && line[len(line)-1] == ' ' && !isEscaped(line, len(line)-1) {
		line = line[:len(line)-1]
	}
	if line == "" {
		return rule{}, false, nil
	}

	// Preserve the normalized rule text before removing matching syntax.
	r := rule{pattern: line}
	if line[0] == '!' {
		r.negated = true
		line = line[1:]
		if line == "" {
			return rule{}, false, nil
		}
	}

	anchored := line[0] == '/'
	if anchored {
		line = line[1:]
	}
	if len(line) > 0 && line[len(line)-1] == '/' && !isEscaped(line, len(line)-1) {
		r.directoryOnly = true
		line = line[:len(line)-1]
	}
	if line == "" {
		return rule{}, false, nil
	}

	// An escaped slash also makes a Git pattern root-relative.
	hasSlash := strings.Contains(line, "/")
	regexSource := `^` + globRegex(line, caseInsensitive) + `$`
	compiled, err := regexp.Compile(regexSource)
	if err != nil {
		return rule{}, false, fmt.Errorf("invalid pattern %q: %w", line, err)
	}
	r.matchBasename = !anchored && !hasSlash
	r.literalPrefix, _ = compiled.LiteralPrefix()
	r.regex = compiled
	return r, true, nil
}

func (r rule) matches(path string, isDir bool) bool {
	if r.directoryOnly && !isDir {
		return false
	}
	if r.matchBasename {
		if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
			path = path[slash+1:]
		}
	}
	path = bytesAsRunes(path)
	if r.literalPrefix != "" && !strings.HasPrefix(path, r.literalPrefix) {
		return false
	}
	return r.regex.MatchString(path)
}

// bytesAsRunes lets RE2 model Git wildmatch's byte-oriented wildcards.
func bytesAsRunes(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			var out strings.Builder
			out.Grow(len(s) * 2)
			for j := 0; j < len(s); j++ {
				out.WriteRune(rune(s[j]))
			}
			return out.String()
		}
	}
	return s
}
