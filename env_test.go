package minijinja_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/maxbrunet/minijinja-go/v2"
)

func TestEnvironment_AddTemplateError(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	err := env.AddTemplate("error", "{% if")
	isTrue(t, err != nil)
	var mjErr *minijinja.Error
	isTrue(t, errors.As(err, &mjErr))
}

func TestEnvironment_RemoveTemplate(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	err := env.AddTemplate("sum", "{{ 1 + 2 }}")
	isEqual(t, nil, err)

	err = env.RemoveTemplate("sum")
	isEqual(t, nil, err)

	_, err = env.RenderTemplate("sum", nil)
	isTrue(t, err != nil)
	var mjErr *minijinja.Error
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, minijinja.ErrorKindTemplateNotFound, mjErr.Kind)
}

func TestEnvironment_ClearTemplates(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	err := env.AddTemplate("sum", "{{ 1 + 2 }}")
	isEqual(t, nil, err)

	err = env.ClearTemplates()
	isEqual(t, nil, err)

	_, err = env.RenderTemplate("sum", nil)
	isTrue(t, err != nil)
	var mjErr *minijinja.Error
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, minijinja.ErrorKindTemplateNotFound, mjErr.Kind)
}

func TestEnvironment_RenderNamedString(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	s, err := env.RenderNamedString("sum", "{{ 1 + 2 + x }}", map[string]int{
		"x": 3,
	})
	isEqual(t, nil, err)
	isEqual(t, "6", s)
}

func TestEnvironment_RenderNamedStringError(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	_, err := env.RenderNamedString("error", "{% if", nil)
	isTrue(t, err != nil)
	var mjErr *minijinja.Error
	isTrue(t, errors.As(err, &mjErr))
}

func TestEnvironment_RenderTemplate(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	err := env.AddTemplate("sum", "{{ 1 + 2 + x }}")
	isEqual(t, nil, err)

	s, err := env.RenderTemplate("sum", map[string]int{
		"x": 3,
	})
	isEqual(t, nil, err)
	isEqual(t, "6", s)
}

func TestEnvironment_RenderTemplateError(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	_, err := env.RenderTemplate("not-found", nil)
	isTrue(t, err != nil)
	var mjErr *minijinja.Error
	isTrue(t, errors.As(err, &mjErr))
}

func TestEnvironment_EvalExpr(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	var res float64
	err := env.EvalExpr("1 + 2 + x", map[string]int{
		"x": 3,
	}, &res)
	isEqual(t, nil, err)
	isEqual(t, 6, res)
}

func TestEnvironment_EvalExprError(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	var res struct{}
	err := env.EvalExpr("{% if", nil, &res)
	isTrue(t, err != nil)
	var mjErr *minijinja.Error
	isTrue(t, errors.As(err, &mjErr))
}

func TestEnvironment_EvalExprInvalid(t *testing.T) {
	env := minijinja.NewEnvironment()
	defer env.Close()

	var res struct{}
	err := env.EvalExpr("{}", nil, res)
	isTrue(t, err != nil)
	var mjErr *minijinja.InvalidEvalExprError
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, reflect.TypeFor[struct{}](), mjErr.Type)
}
