package minijinja_test

import (
	"errors"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/maxbrunet/minijinja-go/v2"
)

func noError(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatal(err)
	}
}

func isEqual[T comparable](tb testing.TB, expected, actual T) {
	tb.Helper()
	if expected != actual {
		tb.Fatalf("expected: %+v\ngot: %+v\n", expected, actual)
	}
}

func inEpsilon(tb testing.TB, expected, actual, epsilon float64) {
	tb.Helper()
	if !(math.Abs(expected-actual) <= epsilon) {
		tb.Fatalf("expected: %+v\ngot: %+v\n", expected, actual)
	}
}

func isTrue(tb testing.TB, value bool) {
	tb.Helper()
	if !value {
		tb.Fatal("should be true")
	}
}

func pointer[T any](v T) *T {
	return &v
}

func testValue[T1, T2 any](t *testing.T, in T1, out *T2) error {
	t.Helper()

	env := minijinja.NewEnvironment()
	defer env.Close()

	return env.EvalExpr("value", map[string]any{"value": in}, out)
}

func mustTestValue[T any](t *testing.T, in T, out *T) {
	t.Helper()

	err := testValue(t, in, out)
	noError(t, err)
}

func TestValue_Bool(t *testing.T) {
	t.Parallel()

	var in, out bool

	in = true
	mustTestValue(t, in, &out)
	isEqual(t, in, out)

	in = false
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_BoolInterface(t *testing.T) {
	t.Parallel()

	var in, out any

	in = true
	mustTestValue(t, in, &out)

	outBool, ok := out.(bool)
	isTrue(t, ok)

	isTrue(t, outBool)
}

func TestValue_BytesArray(t *testing.T) {
	t.Parallel()

	var in, out [3]byte

	in = [3]byte{0x01, 0x02, 0x03}
	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_BytesArrayInterface(t *testing.T) {
	t.Parallel()

	var in []byte
	var out [4]any

	in = []byte{0x01, 0x02, 0x03}
	err := testValue(t, in, &out)
	isEqual(t, nil, err)

	isEqual(t, len(in)+1, len(out))
	for i, e := range in {
		b, ok := out[i].(byte)
		isTrue(t, ok)
		isEqual(t, e, b)
	}
}

func TestValue_BytesArraySmaller(t *testing.T) {
	t.Parallel()

	var in []byte
	var out [3]byte

	in = []byte{0x01, 0x02, 0x03, 0x04}
	err := testValue(t, in, &out)
	isEqual(t, nil, err)
	for i, e := range out {
		isEqual(t, in[i], e)
	}
}

func TestValue_BytesArrayLarger(t *testing.T) {
	t.Parallel()

	in := []byte{0x01, 0x02, 0x03}
	out := [4]byte{0x01, 0x02, 0x03, 0x04}

	err := testValue(t, in, &out)
	isEqual(t, nil, err)
	for i, e := range out {
		if i < len(in) {
			isEqual(t, in[i], e)
			continue
		}
		isEqual(t, 0, e)
	}
}

func TestValue_BytesArrayInterfaceWithMethods2(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := []byte{0x01, 0x2, 0x03}
	var out [4]myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "bytes", mjErr.Value)
	isEqual(t, reflect.TypeFor[[4]myIf](), mjErr.Type)
}

func TestValue_BytesInterface(t *testing.T) {
	t.Parallel()

	var in []byte
	var out []any

	in = []byte{0x01, 0x02, 0x03}
	err := testValue(t, in, &out)
	isEqual(t, nil, err)

	isEqual(t, len(in), len(out))
	for i, e := range in {
		b, ok := out[i].(byte)
		isTrue(t, ok)
		isEqual(t, e, b)
	}
}

func TestValue_BytesInterface2(t *testing.T) {
	t.Parallel()

	var in []byte
	var outI any

	in = []byte{0x01, 0x02, 0x03}
	err := testValue(t, in, &outI)
	isEqual(t, nil, err)

	out, ok := outI.([]byte)
	isTrue(t, ok)

	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_BytesSlice(t *testing.T) {
	t.Parallel()

	var in, out []byte

	in = []byte{0x01, 0x02, 0x03}

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_BytesSliceAllocated(t *testing.T) {
	t.Parallel()

	in := []byte{0x01, 0x02, 0x03}
	out := make([]byte, len(in))

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_BytesSliceAllocatedSmaller(t *testing.T) {
	t.Parallel()

	in := []byte{0x01, 0x02, 0x03}
	out := make([]byte, len(in)-1)

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_BytesSliceAllocatedBigger(t *testing.T) {
	t.Parallel()

	in := []byte{0x01, 0x02, 0x03}
	out := make([]byte, len(in)+1)

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	isEqual(t, cap(in)+1, cap(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_BytesSliceInterfaceWithMethods(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := []byte{}
	var out myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "bytes", mjErr.Value)
	isEqual(t, reflect.TypeFor[myIf](), mjErr.Type)
}

func TestValue_BytesSliceInterfaceWithMethods2(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := []byte{0x01, 0x2, 0x03}
	var out []myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "bytes", mjErr.Value)
	isEqual(t, reflect.TypeFor[[]myIf](), mjErr.Type)
}

func TestValue_NumberFloat32(t *testing.T) {
	t.Parallel()

	var in, out float32

	in = 1.234
	mustTestValue(t, in, &out)
	inEpsilon(t, float64(in), float64(out), 1e-12)
}

func TestValue_NumberFloat64(t *testing.T) {
	t.Parallel()

	var in, out float64

	in = 1.234
	mustTestValue(t, in, &out)
	inEpsilon(t, in, out, 1e-12)
}

func TestValue_NumberInt(t *testing.T) {
	t.Parallel()

	var in, out int

	in = -123
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_NumberInt32(t *testing.T) {
	t.Parallel()

	var in, out int32

	in = -123
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_NumberInt64(t *testing.T) {
	t.Parallel()

	var in, out int64

	in = -123
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_NumberInterface(t *testing.T) {
	t.Parallel()

	var in, out any

	in = 123
	mustTestValue(t, in, &out)

	inInt, ok := in.(int)
	isTrue(t, ok)

	outFloat, ok := out.(float64)
	isTrue(t, ok)

	inEpsilon(t, float64(inInt), outFloat, 1e-12)
}

func TestValue_NumberInterfaceWithMethods(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := 123
	var out myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "number "+strconv.Itoa(in), mjErr.Value)
	isEqual(t, reflect.TypeFor[myIf](), mjErr.Type)
}

func TestValue_NumberUint(t *testing.T) {
	t.Parallel()

	var in, out uint

	in = 123
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_NumberUint32(t *testing.T) {
	t.Parallel()

	var in, out uint32

	in = 123
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_NumberUint64(t *testing.T) {
	t.Parallel()

	var in, out uint64

	in = 123
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_Map(t *testing.T) {
	t.Parallel()

	var in, out map[string]int

	in = map[string]int{"foo": 1, "bar": 2}
	mustTestValue(t, in, &out)

	isEqual(t, len(in), len(out))
	for ki, vi := range in {
		vo, ok := out[ki]
		isTrue(t, ok)

		isEqual(t, vi, vo)
	}
}

func TestValue_MapAllocated(t *testing.T) {
	t.Parallel()

	in := map[string]int{"foo": 1, "bar": 2}
	out := make(map[string]int, len(in))

	mustTestValue(t, in, &out)

	isEqual(t, len(in), len(out))
	for ki, vi := range in {
		vo, ok := out[ki]
		isTrue(t, ok)

		isEqual(t, vi, vo)
	}
}

func TestValue_MapNonEmpty(t *testing.T) {
	t.Parallel()

	in := map[string]int{"foo": 1, "bar": 2}
	out := map[string]int{"foobar": 99, "barfoo": 98}

	mustTestValue(t, in, &out)

	isEqual(t, len(in)+2, len(out))
	for ki, vi := range in {
		vo, ok := out[ki]
		isTrue(t, ok)

		isEqual(t, vi, vo)
	}
}

func TestValue_MapInterface(t *testing.T) {
	t.Parallel()

	var in, out map[any]any

	in = map[any]any{
		"name": "Go",
		"seq":  []any{"First", "Second", 42.0},
		"id":   123.0,
	}
	mustTestValue(t, in, &out)

	isEqual(t, len(in), len(out))
	for ki, vi := range in {
		vo, ok := out[ki]
		isTrue(t, ok)

		if ki == "seq" {
			inSeq, ok := in[ki].([]any)
			isTrue(t, ok)

			outSeq, ok := out[ki].([]any)
			isTrue(t, ok)

			isEqual(t, len(inSeq), len(outSeq))
			for i, e := range inSeq {
				isEqual(t, e, outSeq[i])
			}

			continue
		}

		isEqual(t, vi, vo)
	}
}

func TestValue_MapInterface2(t *testing.T) {
	t.Parallel()

	var in map[any]any
	var outI any

	in = map[any]any{
		"name": "Go",
		"seq":  []any{"First", "Second", 42.0},
		"id":   123.0,
	}
	err := testValue(t, in, &outI)
	isEqual(t, nil, err)

	out, ok := outI.(map[any]any)
	isTrue(t, ok)

	isEqual(t, len(in), len(out))
	for ki, vi := range in {
		vo, ok := out[ki]
		isTrue(t, ok)

		if ki == "seq" {
			inSeq, ok := in[ki].([]any)
			isTrue(t, ok)

			outSeq, ok := out[ki].([]any)
			isTrue(t, ok)

			isEqual(t, len(inSeq), len(outSeq))
			for i, e := range inSeq {
				isEqual(t, e, outSeq[i])
			}

			continue
		}

		isEqual(t, vi, vo)
	}
}

func TestValue_MapInterfaceWithMethods(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := map[any]any{}
	var out myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "map", mjErr.Value)
	isEqual(t, reflect.TypeFor[myIf](), mjErr.Type)
}

func TestValue_NonePointer(t *testing.T) {
	t.Parallel()

	var in, out *string

	in = nil
	out = pointer("not-nil")
	mustTestValue(t, in, &out)
	isTrue(t, out != nil)
	isEqual(t, "not-nil", *out)
}

func TestValue_NoneNonPointer(t *testing.T) {
	t.Parallel()

	var in *int
	out := 3

	err := testValue(t, in, &out)
	isEqual(t, nil, err)
	isEqual(t, 3, out)
}

func TestValue_NoneInterface(t *testing.T) {
	t.Parallel()

	var in, out any

	in = nil
	out = "not-nil"
	mustTestValue(t, in, &out)
	isEqual(t, "not-nil", out)
}

func TestValue_NoneInterfaceWithMethods(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	var in any
	var out myIf

	err := testValue(t, in, &out)
	isEqual(t, nil, err)
	isEqual(t, nil, out)
}

func TestValue_NoneMap(t *testing.T) {
	t.Parallel()

	var in, out map[any]any

	in = nil
	out = map[any]any{}
	mustTestValue(t, in, &out)

	isTrue(t, out != nil)
}

func TestValue_NoneSlice(t *testing.T) {
	t.Parallel()

	var in, out []any

	in = nil
	out = []any{}
	mustTestValue(t, in, &out)

	isTrue(t, out != nil)
}

func TestValue_Plain(t *testing.T) {
	t.Parallel()

	t.Helper()

	env := minijinja.NewEnvironment()
	defer env.Close()

	var out string

	err := env.EvalExpr("debug", nil, &out)
	isEqual(t, nil, err)
	isEqual(t, "minijinja::functions::builtins::debug", out)
}

func TestValue_PlainInterface(t *testing.T) {
	t.Parallel()

	t.Helper()

	env := minijinja.NewEnvironment()
	defer env.Close()

	var out any

	err := env.EvalExpr("debug", nil, &out)
	isEqual(t, nil, err)
	isEqual(t, "minijinja::functions::builtins::debug", out)
}

func TestValue_Pointer(t *testing.T) {
	t.Parallel()

	var in, out *string

	in = pointer("foobar")
	out = nil
	mustTestValue(t, in, &out)
	isTrue(t, out != nil)
	isEqual(t, *in, *out)
}

func TestValue_Pointer2(t *testing.T) {
	t.Parallel()

	var in, out *struct{}

	in = &struct{}{}
	out = nil
	mustTestValue(t, in, &out)
	isTrue(t, out != nil)
	isEqual(t, *in, *out)
}

func TestValue_String(t *testing.T) {
	t.Parallel()

	var in, out string

	in = "foobar"
	mustTestValue(t, in, &out)
	isEqual(t, in, out)
}

func TestValue_StringInterface(t *testing.T) {
	t.Parallel()

	var in, out any

	in = "foobar-if"
	mustTestValue(t, in, &out)

	inStr, ok := in.(string)
	isTrue(t, ok)

	outStr, ok := out.(string)
	isTrue(t, ok)

	isEqual(t, inStr, outStr)
}

func TestValue_StringInterfaceWithMethods(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := "foobar-if-methods"
	var out myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "string", mjErr.Value)
	isEqual(t, reflect.TypeFor[myIf](), mjErr.Type)
}

func TestValue_SeqArray(t *testing.T) {
	t.Parallel()

	var in, out [3]string

	in = [3]string{"First", "Second", "Third"}
	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SeqArraySmaller(t *testing.T) {
	t.Parallel()

	var in []string
	var out [3]string

	in = []string{"First", "Second", "Third", "Four"}
	err := testValue(t, in, &out)
	isEqual(t, nil, err)
	for i, e := range out {
		isEqual(t, in[i], e)
	}
}

func TestValue_SeqArrayLarger(t *testing.T) {
	t.Parallel()

	in := []string{"First", "Second", "Third"}
	out := [4]string{"1", "2", "3", "4"}

	err := testValue(t, in, &out)
	isEqual(t, nil, err)
	for i, e := range out {
		if i < len(in) {
			isEqual(t, in[i], e)
			continue
		}
		isEqual(t, "", e)
	}
}

func TestValue_SeqSlice(t *testing.T) {
	t.Parallel()

	var in, out []string

	in = []string{"First", "Second", "Third"}

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SeqSliceAllocated(t *testing.T) {
	t.Parallel()

	in := []string{"First", "Second", "Third"}
	out := make([]string, len(in))

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SeqSliceAllocatedSmaller(t *testing.T) {
	t.Parallel()

	in := []string{"First", "Second", "Third"}
	out := make([]string, len(in)-1)

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SeqSliceAllocatedBigger(t *testing.T) {
	t.Parallel()

	in := []string{"First", "Second", "Third"}
	out := make([]string, len(in)+1)

	mustTestValue(t, in, &out)
	isEqual(t, len(in), len(out))
	isEqual(t, cap(in)+1, cap(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SeqInterface(t *testing.T) {
	t.Parallel()

	var in, out []any

	in = []any{"First", "Second", 42.0}
	mustTestValue(t, in, &out)

	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SeqInterface2(t *testing.T) {
	t.Parallel()

	var in []any
	var outI any

	in = []any{"First", "Second", 42.0}
	err := testValue(t, in, &outI)
	isEqual(t, nil, err)

	out, ok := outI.([]any)
	isTrue(t, ok)

	isEqual(t, len(in), len(out))
	for i, e := range in {
		isEqual(t, e, out[i])
	}
}

func TestValue_SliceInterfaceWithMethods(t *testing.T) {
	t.Parallel()

	type myIf interface{ DoStuff() }
	in := []any{}
	var out myIf

	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "seq", mjErr.Value)
	isEqual(t, reflect.TypeFor[myIf](), mjErr.Type)
}

func TestValue_Struct(t *testing.T) {
	t.Parallel()

	type thisStruct struct {
		Name string `minijinja:"name"`
		Seq  []any  `minijinja:"seq"`
		ID   int    `minijinja:"id"`
	}

	var in, out thisStruct

	in = thisStruct{
		Name: "Go",
		Seq:  []any{"First", "Second", 42.0},
		ID:   123,
	}
	mustTestValue(t, in, &out)
	isEqual(t, in.Name, out.Name)
	isEqual(t, len(in.Seq), len(out.Seq))
	for i, e := range in.Seq {
		isEqual(t, e, out.Seq[i])
	}
	isEqual(t, in.ID, out.ID)
}

func TestValue_NotMap(t *testing.T) {
	t.Parallel()

	var in int
	var out map[int]int

	in = 1
	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "number "+strconv.Itoa(in), mjErr.Value)
	isEqual(t, reflect.TypeFor[map[int]int](), mjErr.Type)
}

func TestValue_NotNumber(t *testing.T) {
	t.Parallel()

	var in string
	var out int

	in = "1"
	err := testValue(t, in, &out)
	isTrue(t, err != nil)
	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "string", mjErr.Value)
	isEqual(t, reflect.TypeFor[int](), mjErr.Type)
}

func TestValue_NotSlice(t *testing.T) {
	t.Parallel()

	var in int
	var out []int

	in = 1
	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "number "+strconv.Itoa(in), mjErr.Value)
	isEqual(t, reflect.TypeFor[[]int](), mjErr.Type)
}

func TestValue_NotSlice2(t *testing.T) {
	t.Parallel()

	in := map[string]string{"foo": "bar"}
	var out []string

	err := testValue(t, in, &out)
	isTrue(t, err != nil)
	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "map", mjErr.Value)
	isEqual(t, reflect.TypeFor[[]string](), mjErr.Type)
}

func TestValue_NotString(t *testing.T) {
	t.Parallel()

	var in int
	var out string

	in = 1
	err := testValue(t, in, &out)
	isTrue(t, err != nil)
	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "number "+strconv.Itoa(in), mjErr.Value)
	isEqual(t, reflect.TypeFor[string](), mjErr.Type)
}

func TestValue_NotStruct(t *testing.T) {
	t.Parallel()

	var in int
	var out struct{}

	in = 1
	err := testValue(t, in, &out)
	isTrue(t, err != nil)

	mjErr := &minijinja.DecodeTypeError{}
	isTrue(t, errors.As(err, &mjErr))
	isEqual(t, "number "+strconv.Itoa(in), mjErr.Value)
	isEqual(t, reflect.TypeFor[struct{}](), mjErr.Type)
}
