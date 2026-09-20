package gitignorematch

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkMatch(b *testing.B) {
	for _, ruleCount := range []int{10, 100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("rules=%d", ruleCount), func(b *testing.B) {
			var patterns strings.Builder
			for i := 0; i < ruleCount; i++ {
				fmt.Fprintf(&patterns, "a/b/c/target-%05d\n", i)
			}
			matcher, err := Parse(patterns.String())
			if err != nil {
				b.Fatal(err)
			}
			tests := []struct {
				name string
				path string
				want Decision
			}{
				{
					name: "last-rule-match",
					path: fmt.Sprintf("a/b/c/target-%05d", ruleCount-1),
					want: Ignore,
				},
				{name: "no-match", path: "a/b/c/not-present", want: NoMatch},
			}
			for _, tt := range tests {
				b.Run(tt.name, func(b *testing.B) {
					b.ReportAllocs()
					var got Decision
					for i := 0; i < b.N; i++ {
						got = matcher.Match(tt.path, false)
					}
					if got != tt.want {
						b.Fatalf("Match() = %s, want %s", got, tt.want)
					}
				})
			}
		})
	}
}
