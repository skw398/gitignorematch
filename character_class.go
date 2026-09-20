package gitignorematch

import (
	"fmt"
	"strings"
)

func globClass(pattern string, start int, caseInsensitive bool) (string, int) {
	i := start + 1
	if i < len(pattern) && (pattern[i] == '!' || pattern[i] == '^') {
		i++
	}
	// A closing bracket is literal when it is the first class member.
	if i < len(pattern) && pattern[i] == ']' {
		i++
	}
	for i < len(pattern) {
		b := pattern[i]
		if b == '\\' {
			i++
			if i < len(pattern) {
				i++
			}
			continue
		}
		if end, ok := posixClassEnd(pattern, i); ok {
			i = end
			continue
		}
		if b == ']' {
			content := pattern[start+1 : i]
			class, valid := classRegex(content, caseInsensitive)
			if !valid {
				return neverMatchRegex, i + 1
			}
			return class, i + 1
		}
		i++
	}
	return neverMatchRegex, start + 1
}

func posixClassEnd(pattern string, start int) (int, bool) {
	if start+1 >= len(pattern) || pattern[start] != '[' || pattern[start+1] != ':' {
		return 0, false
	}
	end := strings.Index(pattern[start+2:], ":]")
	if end < 0 {
		return 0, false
	}
	return start + 2 + end + 2, true
}

func classRegex(content string, caseInsensitive bool) (string, bool) {
	var members [256]bool
	negated := false
	if strings.HasPrefix(content, "!") || strings.HasPrefix(content, "^") {
		negated = true
		content = content[1:]
	}

	previous := -1
	for i := 0; i < len(content); {
		if end, ok := posixClassEnd(content, i); ok {
			if !addPOSIXClass(&members, content[i+2:end-2], caseInsensitive) {
				return "", false
			}
			previous = -1
			i = end
			continue
		}

		b := content[i]
		if b == '\\' && i+1 < len(content) {
			b = content[i+1]
			members[b] = true
			previous = int(b)
			i += 2
			continue
		}
		if b == '-' && previous >= 0 && i+1 < len(content) {
			end := content[i+1]
			consumed := 2
			if end == '\\' && i+2 < len(content) {
				end = content[i+2]
				consumed = 3
			}
			for value := previous; value <= int(end); value++ {
				members[value] = true
			}
			previous = -1
			i += consumed
			continue
		}
		members[b] = true
		previous = int(b)
		i++
	}

	if caseInsensitive {
		expandASCIICase(&members)
	}
	if negated {
		for i := range members {
			members[i] = !members[i]
		}
	}
	// Wildmatch never lets a bracket expression consume a path separator.
	members['/'] = false
	return byteSetRegex(members), true
}

func expandASCIICase(members *[256]bool) {
	for b := byte('A'); b <= 'Z'; b++ {
		if members[b] || members[b+'a'-'A'] {
			members[b] = true
			members[b+'a'-'A'] = true
		}
	}
}

func addPOSIXClass(members *[256]bool, name string, caseInsensitive bool) bool {
	for value := 0; value < len(members); value++ {
		b := byte(value)
		alpha := (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
		digit := b >= '0' && b <= '9'
		var match bool
		switch name {
		case "alnum":
			match = alpha || digit
		case "alpha":
			match = alpha
		case "blank":
			match = b == ' ' || b == '\t'
		case "cntrl":
			match = b < ' ' || b == 0x7f
		case "digit":
			match = digit
		case "graph":
			match = b >= '!' && b <= '~'
		case "lower", "upper":
			if caseInsensitive {
				match = alpha
			} else if name == "lower" {
				match = b >= 'a' && b <= 'z'
			} else {
				match = b >= 'A' && b <= 'Z'
			}
		case "print":
			match = b >= ' ' && b <= '~'
		case "punct":
			match = b >= '!' && b <= '~' && !alpha && !digit
		case "space":
			// Git excludes vertical tab and form feed from its space class.
			match = b == ' ' || b == '\t' || b == '\n' || b == '\r'
		case "xdigit":
			match = digit || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
		default:
			return false
		}
		if match {
			members[value] = true
		}
	}
	return true
}

func byteSetRegex(members [256]bool) string {
	var out strings.Builder
	out.WriteByte('[')
	for start := 0; start < len(members); {
		if !members[start] {
			start++
			continue
		}
		end := start
		for end+1 < len(members) && members[end+1] {
			end++
		}
		fmt.Fprintf(&out, `\x%02X`, start)
		if end != start {
			fmt.Fprintf(&out, `-\x%02X`, end)
		}
		start = end + 1
	}
	if out.Len() == 1 {
		return neverMatchRegex
	}
	out.WriteByte(']')
	return out.String()
}
