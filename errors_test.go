package minijinja_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/maxbrunet/minijinja-go/v2"
)

func TestError(t *testing.T) {
	t.Parallel()

	env := minijinja.NewEnvironment()
	defer env.Close()

	err := env.AddTemplate("hello", "{{ foo | bad_filter }}")
	isEqual(t, nil, err)

	_, err = env.RenderTemplate("hello", nil)
	isTrue(t, err != nil)

	mjErr := &minijinja.Error{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, minijinja.ErrorKindUnknownFilter, mjErr.Kind)
	isEqual(t, "filter bad_filter is unknown", mjErr.Detail)
	isEqual(t, "hello", mjErr.Name)
	isEqual(t, 1, mjErr.Line)
	isEqual(
		t,
		fmt.Sprintf(
			"minijinja: %s: %s (in %s:%d)",
			mjErr.Kind,
			mjErr.Detail,
			mjErr.Name,
			mjErr.Line,
		),
		mjErr.Error(),
	)
}

func TestErrorWithDebug(t *testing.T) {
	t.Parallel()

	env := minijinja.NewEnvironment()
	defer env.Close()
	env.SetDebug(true)

	err := env.AddTemplate("hello", "{{ foo | bad_filter }}")
	isEqual(t, nil, err)

	_, err = env.RenderTemplate("hello", nil)
	isTrue(t, err != nil)

	mjErr := &minijinja.Error{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, minijinja.ErrorKindUnknownFilter, mjErr.Kind)
	isEqual(t, "filter bad_filter is unknown", mjErr.Detail)
	isEqual(t, "hello", mjErr.Name)
	isEqual(t, 1, mjErr.Line)
	isTrue(
		t,
		strings.HasPrefix(
			mjErr.Error(),
			fmt.Sprintf(
				"minijinja: %s: %s (in %s:%d)\n\n---",
				mjErr.Kind,
				mjErr.Detail,
				mjErr.Name,
				mjErr.Line,
			),
		),
	)
}

func TestErrorWithoutName(t *testing.T) {
	t.Parallel()

	env := minijinja.NewEnvironment()
	defer env.Close()

	_, err := env.RenderTemplate("hello", nil)
	isTrue(t, err != nil)
	mjErr := &minijinja.Error{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, minijinja.ErrorKindTemplateNotFound, mjErr.Kind)
	isEqual(t, "template \"hello\" does not exist", mjErr.Detail)
	isEqual(t, "", mjErr.Name)
	isEqual(t, 0, mjErr.Line)
	isEqual(
		t,
		fmt.Sprintf("minijinja: %s: %s", mjErr.Kind, mjErr.Detail),
		mjErr.Error(),
	)
}
