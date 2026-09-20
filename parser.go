package gitignorematch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Options configures pattern matching.
type Options struct {
	// CaseInsensitive makes ASCII letters match without regard to case, like
	// Git with core.ignoreCase enabled. The caller selects this behavior; the
	// package does not read Git configuration.
	CaseInsensitive bool
}

// Parse parses newline-separated Git ignore patterns with default options.
func Parse(patterns string) (*Matcher, error) {
	return ParseWithOptions(patterns, Options{})
}

// ParseWithOptions parses newline-separated Git ignore patterns with options.
func ParseWithOptions(patterns string, options Options) (*Matcher, error) {
	return ParseReaderWithOptions(strings.NewReader(patterns), options)
}

// ParseReader parses Git ignore patterns from r with default options.
func ParseReader(r io.Reader) (*Matcher, error) {
	return ParseReaderWithOptions(r, Options{})
}

// ParseReaderWithOptions parses Git ignore patterns from r with options.
func ParseReaderWithOptions(r io.Reader, options Options) (*Matcher, error) {
	if r == nil {
		return nil, fmt.Errorf("gitignorematch: nil reader")
	}

	m := &Matcher{caseInsensitive: options.CaseInsensitive}
	reader := bufio.NewReader(r)
	lineNumber := 0
	for {
		line, err := reader.ReadString('\n')
		if len(line) != 0 {
			lineNumber++
			line, _, _ = strings.Cut(line, "\x00")
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			if lineNumber == 1 {
				line = strings.TrimPrefix(line, "\uFEFF")
			}

			parsedRule, keep, parseErr := parseRule(line, options.CaseInsensitive)
			if parseErr != nil {
				return nil, fmt.Errorf("gitignorematch: line %d: %w", lineNumber, parseErr)
			}
			if keep {
				parsedRule.line = lineNumber
				m.rules = append(m.rules, parsedRule)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("gitignorematch: read rules: %w", err)
		}
	}
	return m, nil
}

// ParseFile parses Git ignore patterns from name with default options.
func ParseFile(name string) (*Matcher, error) {
	return ParseFileWithOptions(name, Options{})
}

// ParseFileWithOptions parses Git ignore patterns from name with options.
func ParseFileWithOptions(name string, options Options) (*Matcher, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ParseReaderWithOptions(file, options)
}
