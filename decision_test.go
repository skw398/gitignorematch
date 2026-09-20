package gitignorematch

import "testing"

func TestDecisionString(t *testing.T) {
	tests := []struct {
		decision Decision
		want     string
	}{
		{decision: NoMatch, want: "NoMatch"},
		{decision: Ignore, want: "Ignore"},
		{decision: Include, want: "Include"},
		{decision: Decision(255), want: "NoMatch"},
	}
	for _, tt := range tests {
		if got := tt.decision.String(); got != tt.want {
			t.Errorf("Decision(%d).String() = %q, want %q", tt.decision, got, tt.want)
		}
	}
}
