package minijinja_test

import (
	"fmt"

	"github.com/maxbrunet/minijinja-go/v2"
)

func ExampleNewEnvironment() {
	env := minijinja.NewEnvironment()
	if env == nil {
		panic("could not create environment")
	}
	defer env.Close()
	env.SetDebug(true)

	templateSource := `Hello {{ name }}!
{%- for item in seq %}
  - {{ item }}
{%- endfor %}
seq: {{ seq }}`
	if err := env.AddTemplate("hello", templateSource); err != nil {
		panic(err)
	}

	ctx := struct {
		Name string `minijinja:"name"`
		Seq  []any  `minijinja:"seq"`
	}{
		Name: "Go",
		Seq:  []any{"First", "Second", 42},
	}

	// render a template
	rendered, err := env.RenderTemplate("hello", ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(rendered)

	// eval an expression
	var res int
	if err := env.EvalExpr("1 + 2", nil, &res); err != nil {
		panic(err)
	}
	fmt.Println("1 + 2 =", res)
	// Output:
	// Hello Go!
	//   - First
	//   - Second
	//   - 42
	// seq: ["First", "Second", 42]
	// 1 + 2 = 3
}
