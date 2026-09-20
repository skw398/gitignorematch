package gitignorematch

import "fmt"

func Example() {
	matcher, err := Parse("*.log\n!important.log\n")
	if err != nil {
		panic(err)
	}

	fmt.Println(matcher.Match("app.log", false))
	fmt.Println(matcher.Explain("important.log", false).Line)

	// Output:
	// Ignore
	// 2
}
