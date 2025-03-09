package minijinja_test

import (
	"testing"

	"github.com/maxbrunet/minijinja-go/v2"
)

func TestSynthaxConfig(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	err := env.SetSyntaxConfig(&minijinja.SyntaxConfig{
		BlockStart:          "%%",
		BlockEnd:            "%%",
		VariableStart:       "<<",
		VariableEnd:         ">>",
		CommentStart:        "/*",
		CommentEnd:          "*/",
		LineStatementPrefix: "#",
		LineCommentPrefix:   "//",
	})
	isEqual(t, nil, err)

	res, err := env.RenderNamedString("syntax",
		`// A comment
# if true
A custom syntax config
# endif
%%- for item in seq %%
- << item >>
%%- endfor %%`, map[string][]string{"seq": {"foo", "bar"}})
	isEqual(t, nil, err)
	isEqual(t, `A custom syntax config

- foo
- bar`, res)
}
