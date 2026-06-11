package a2s_test

import (
	"fmt"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func ExampleNew() {
	client, err := a2s.New(
		"127.0.0.1",
		27015,
		a2s.WithTimeout(3*time.Second),
	)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Println(client.Addr())
	// Output: 127.0.0.1:27015
}

func ExampleParseRuleValues() {
	rules := a2s.ParseRuleValues(a2s.Rules{
		{Name: "hostname", Value: "Example server"},
		{Name: "players", Value: "12"},
		{Name: "secure", Value: "true"},
	})

	fmt.Printf("%s %T %T\n", rules["hostname"], rules["players"], rules["secure"])
	// Output: Example server int64 bool
}
