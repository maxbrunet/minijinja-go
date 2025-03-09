package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #include <stdlib.h>
// #include <minijinja.h>
import "C"

import (
	"fmt"
	"reflect"
	"strconv"
)

// ErrorKind represents the kind of error that occurred.
type ErrorKind uint

const (
	// ErrorKindNonPrimitive indicates a non primitive value was encountered
	// where one was expected.
	ErrorKindNonPrimitive ErrorKind = iota
	// ErrorKindNonKey indicates a value is not valid for a key in a map.
	ErrorKindNonKey
	// ErrorKindInvalidOperation indicates an invalid operation was attempted.
	ErrorKindInvalidOperation
	// ErrorKindSyntaxError indicates the template has a syntax error.
	ErrorKindSyntaxError
	// ErrorKindTemplateNotFound indicates a template was not found.
	ErrorKindTemplateNotFound
	// ErrorKindTooManyArguments indicates too many arguments were passed to a
	// function.
	ErrorKindTooManyArguments
	// ErrorKindMissingArgument indicates an expected argument was missing.
	ErrorKindMissingArgument
	// ErrorKindUnknownFilter indicates a filter is unknown.
	ErrorKindUnknownFilter
	// ErrorKindUnknownFunction indicates a function is unknown.
	ErrorKindUnknownFunction
	// ErrorKindUnknownTest indicates a test is unknown.
	ErrorKindUnknownTest
	// ErrorKindUnknownMethod indicates an unknown method was called.
	ErrorKindUnknownMethod
	// ErrorKindBadEscape indicates a bad escape sequence in a string was
	// encountered.
	ErrorKindBadEscape
	// ErrorKindUndefinedError indicates an operation on an undefined value was
	// attempted.
	ErrorKindUndefinedError
	// ErrorKindBadSerialization indicates not able to serialize this value.
	ErrorKindBadSerialization
	// ErrorKindBadInclude indicates an error happened in an include.
	ErrorKindBadInclude
	// ErrorKindEvalBlock indicates an error happened in a super block.
	ErrorKindEvalBlock
	// ErrorKindCannotUnpack indicates unable to unpack a value.
	ErrorKindCannotUnpack
	// ErrorKindWriteFailure indicates failed writing output.
	ErrorKindWriteFailure
	// ErrorKindUnknown indicates an unknown block was called.
	ErrorKindUnknown
)

var errorKinds = []string{
	ErrorKindNonPrimitive:     "not a primitive",
	ErrorKindNonKey:           "not a key type",
	ErrorKindInvalidOperation: "invalid operation",
	ErrorKindSyntaxError:      "syntax error",
	ErrorKindTemplateNotFound: "template not found",
	ErrorKindTooManyArguments: "too many arguments",
	ErrorKindMissingArgument:  "missing argument",
	ErrorKindUnknownFilter:    "unknown filter",
	ErrorKindUnknownFunction:  "unknown function",
	ErrorKindUnknownTest:      "unknown test",
	ErrorKindUnknownMethod:    "unknown method",
	ErrorKindBadEscape:        "bad string escape",
	ErrorKindUndefinedError:   "undefined value",
	ErrorKindBadSerialization: "could not serialize to value",
	ErrorKindBadInclude:       "could not render include",
	ErrorKindEvalBlock:        "could not render block",
	ErrorKindCannotUnpack:     "cannot unpack",
	ErrorKindWriteFailure:     "failed to write output",
	ErrorKindUnknown:          "unknown error",
}

func (e ErrorKind) String() string {
	if uint(e) < uint(len(errorKinds)) {
		return errorKinds[uint(e)]
	}
	return "errorKind" + strconv.Itoa(int(e))
}

// Error represents template errors.
type Error struct {
	// Error kind.
	Kind ErrorKind
	// The detail is an error message that provides further details about the
	// error kind.
	Detail string
	// The filename of the template that caused the error.
	Name string
	// The line number where the error occurred.
	Line uint32
}

func getError() *Error {
	defer clearError()
	return &Error{
		Kind:   ErrorKind(C.mj_err_get_kind()),
		Detail: C.GoString(C.mj_err_get_detail()),
		Name:   C.GoString(C.mj_err_get_template_name()),
		Line:   uint32(C.mj_err_get_line()),
	}
}

func (e *Error) Error() string {
	if e.Kind == ErrorKindUnknown {
		return "minijinja: " + e.Kind.String()
	}

	if e.Name != "" {
		return fmt.Sprintf(
			"minijinja: %s: %s (in %s:%d)",
			e.Kind,
			e.Detail,
			e.Name,
			e.Line,
		)
	}

	return fmt.Sprintf("minijinja: %s: %s", e.Kind, e.Detail)
}

func isErrorSet() bool {
	return bool(C.mj_err_is_set())
}

func clearError() {
	C.mj_err_clear()
}

// A DecodeTypeError describes a minijinja value that was not appropriate for
// a value of a specific Go type.
type DecodeTypeError struct {
	// description of minijinja value - "bool", "seq", "number -5"
	Value string
	// type of Go value it could not be assigned to
	Type reflect.Type
}

func (e *DecodeTypeError) Error() string {
	return fmt.Sprintf(
		"minijinja: cannot decode %s into Go value of type %s",
		e.Value,
		e.Type.String(),
	)
}

// An InvalidEvalExprError describes an invalid argument passed to [EvalExpr].
// (The argument to [EvalExpr] must be a non-nil pointer.)
type InvalidEvalExprError struct {
	Type reflect.Type
}

func (e *InvalidEvalExprError) Error() string {
	if e.Type == nil {
		return "minijinja: EvalExpr(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return fmt.Sprintf("minijinja: EvalExpr(non-pointer %s)", e.Type)
	}

	return fmt.Sprintf("minijinja: EvalExpr(nil %s)", e.Type)
}

// A UnsupportedTypeError is returned when attempting to encode an unsupported
// value type.
type UnsupportedTypeError struct {
	// type of Go value
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "minijinja: cannot encode Go value of unsupported type " +
		e.Type.String()
}
