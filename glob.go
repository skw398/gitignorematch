package gitignorematch

import (
	"regexp"
	"strings"
)

// globRegex translates the Git wildmatch subset used by .gitignore into a
// regular expression. The caller adds path anchors and basename matching.
func globRegex(pattern string, caseInsensitive bool) string {
	// A trailing /** has a distinct Git meaning: everything inside the
	// directory, but not the directory itself.
	if len(pattern) >= 3 && strings.HasSuffix(pattern, "/**") &&
		!isEscaped(pattern, len(pattern)-3) {
		return globFragment(pattern[:len(pattern)-3], caseInsensitive) + `/(?s:.+)`
	}
	return globFragment(pattern, caseInsensitive)
}

// neverMatchRegex requires a character after the end of the input.
const neverMatchRegex = `\z(?s:.)`

func globFragment(pattern string, caseInsensitive bool) string {
	var out strings.Builder
	for i := 0; i < len(pattern); {
		b := pattern[i]
		switch b {
		case '\\':
			if i+1 >= len(pattern) {
				out.WriteString(neverMatchRegex)
				i++
				continue
			}
			literal := pattern[i+1]
			if caseInsensitive {
				literal = foldASCIIByte(literal)
			}
			out.WriteString(regexp.QuoteMeta(string(rune(literal))))
			i += 2
		case '*':
			start := i
			for i < len(pattern) && pattern[i] == '*' {
				i++
			}
			count := i - start
			componentStart := start == 0 || pattern[start-1] == '/'
			if count >= 2 && componentStart && i < len(pattern) && pattern[i] == '/' {
				out.WriteString(`(?:[^/]+/)*`)
				i++
			} else if count >= 2 && componentStart && i+1 < len(pattern) &&
				pattern[i] == '\\' && pattern[i+1] == '/' {
				out.WriteString(`(?s:.*)`)
			} else {
				out.WriteString(`[^/]*`)
			}
		case '?':
			out.WriteString(`[^/]`)
			i++
		case '[':
			class, next := globClass(pattern, i, caseInsensitive)
			out.WriteString(class)
			i = next
		default:
			if caseInsensitive {
				b = foldASCIIByte(b)
			}
			out.WriteString(regexp.QuoteMeta(string(rune(b))))
			i++
		}
	}
	return out.String()
}

func foldASCII(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			var out strings.Builder
			out.Grow(len(s))
			out.WriteString(s[:i])
			for ; i < len(s); i++ {
				out.WriteByte(foldASCIIByte(s[i]))
			}
			return out.String()
		}
	}
	return s
}

func foldASCIIByte(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 'a' - 'A'
	}
	return b
}

func isEscaped(s string, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && s[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}
