package gitignorematch

import (
	"path"
	"strings"
)

// Matcher is an immutable, reusable set of parsed .gitignore rules.
type Matcher struct {
	rules           []rule
	caseInsensitive bool
}

// Match returns the effective decision for path. Paths are slash-separated
// and relative to the root at which these rules were read. An absolute path
// or a path that escapes that root returns NoMatch. A directory rule also
// applies to descendants of that directory, which is necessary to model
// Git's inability to traverse an ignored parent directory.
func (m *Matcher) Match(path string, isDir bool) Decision {
	decision, _ := m.match(path, isDir)
	return decision
}

// Explain returns the effective decision and the rule that determined it.
// Pattern omits unescaped trailing spaces and Line is one-based. For NoMatch,
// Pattern is empty and Line is zero.
func (m *Matcher) Explain(path string, isDir bool) MatchResult {
	decision, matchedRule := m.match(path, isDir)
	if matchedRule == nil {
		return MatchResult{Decision: decision}
	}
	return MatchResult{
		Decision: decision,
		Pattern:  matchedRule.pattern,
		Line:     matchedRule.line,
		Negated:  matchedRule.negated,
	}
}

func (m *Matcher) match(path string, isDir bool) (Decision, *rule) {
	if m == nil {
		return NoMatch, nil
	}
	p := cleanPath(path)
	if p == "" {
		return NoMatch, nil
	}
	if m.caseInsensitive {
		p = foldASCII(p)
	}

	// An ignored parent prevents later rules from including a descendant.
	for end := strings.IndexByte(p, '/'); end >= 0; {
		if decision, matchedRule := m.decisionFor(p[:end], true); decision == Ignore {
			return decision, matchedRule
		}
		next := strings.IndexByte(p[end+1:], '/')
		if next < 0 {
			break
		}
		end += next + 1
	}
	return m.decisionFor(p, isDir)
}

func (m *Matcher) decisionFor(path string, isDir bool) (Decision, *rule) {
	for i := len(m.rules) - 1; i >= 0; i-- {
		candidate := &m.rules[i]
		if !candidate.matches(path, isDir) {
			continue
		}
		if candidate.negated {
			return Include, candidate
		}
		return Ignore, candidate
	}
	return NoMatch, nil
}

func cleanPath(p string) string {
	if strings.HasPrefix(p, "/") {
		return ""
	}
	p = path.Clean(p)
	if p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return ""
	}
	return p
}
