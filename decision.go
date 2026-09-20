package gitignorematch

// Decision is the result of matching a path against the rules.
type Decision uint8

const (
	// NoMatch means that no effective rule matched the path.
	NoMatch Decision = iota
	// Ignore means that the path is ignored by the rules.
	Ignore
	// Include means that a negated rule explicitly includes the path.
	Include
)

// String returns a readable name for d.
func (d Decision) String() string {
	switch d {
	case Ignore:
		return "Ignore"
	case Include:
		return "Include"
	default:
		return "NoMatch"
	}
}

// MatchResult explains the effective decision for a path. Pattern and Line
// identify the rule that determined Decision. Pattern omits unescaped trailing
// spaces. For NoMatch, Pattern is empty and Line is zero.
type MatchResult struct {
	Decision Decision
	Pattern  string
	Line     int
	Negated  bool
}
