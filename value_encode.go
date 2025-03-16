package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #cgo nocallback mj_value_new_bool
// #cgo nocallback mj_value_new_f32
// #cgo nocallback mj_value_new_f64
// #cgo nocallback mj_value_new_i32
// #cgo nocallback mj_value_new_i64
// #cgo nocallback mj_value_new_list
// #cgo nocallback mj_value_new_none
// #cgo nocallback mj_value_new_object
// #cgo noescape mj_value_new_string
// #cgo nocallback mj_value_new_string
// #cgo nocallback mj_value_new_u32
// #cgo nocallback mj_value_new_u64
// #include <stdlib.h>
// #include <minijinja.h>
import "C"

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

// newValue creates a new value.
func newValue(x any) (*value, error) {
	switch x := x.(type) {
	case bool:
		return newValueBool(x), nil
	case float32:
		return newValueFloat32(x), nil
	case float64:
		return newValueFloat64(x), nil
	case int:
		return newValueInt(x), nil
	case int32:
		return newValueInt32(x), nil
	case int64:
		return newValueInt64(x), nil
	case nil:
		return newValueNone(), nil
	case string:
		return newValueString(x), nil
	case uint:
		return newValueUint(x), nil
	case uint32:
		return newValueUint32(x), nil
	case uint64:
		return newValueUint64(x), nil
	}

	rv := reflect.ValueOf(x)
	switch rv.Kind() {
	case reflect.Array:
		return newValueSeq(x)
	case reflect.Map:
		if rv.IsNil() {
			return newValueNone(), nil
		}
		return newValueMap(x)
	case reflect.Pointer:
		if rv.IsNil() {
			return newValueNone(), nil
		}
		return newValue(rv.Elem().Interface())
	case reflect.Struct:
		return newValueStruct(x)
	case reflect.Slice:
		if rv.IsNil() {
			return newValueNone(), nil
		}
		return newValueSeq(x)
	}

	return nil, &UnsupportedTypeError{Type: rv.Type()}
}

func newValueBool(x bool) *value {
	return &value{cVal: C.mj_value_new_bool(C.bool(x))}
}

func newValueFloat32(x float32) *value {
	return &value{cVal: C.mj_value_new_f32(C.float(x))}
}

func newValueFloat64(x float64) *value {
	return &value{cVal: C.mj_value_new_f64(C.double(x))}
}

func newValueInt(x int) *value {
	switch math.MaxInt {
	case math.MaxInt32:
		return newValueInt32(int32(x))
	case math.MaxInt64:
		return newValueInt64(int64(x))
	}

	panic(fmt.Sprintf("unexpected max int: %d", math.MaxInt))
}

func newValueInt32(x int32) *value {
	return &value{cVal: C.mj_value_new_i32(C.int32_t(x))}
}

func newValueInt64(x int64) *value {
	return &value{cVal: C.mj_value_new_i64(C.int64_t(x))}
}

func newValueMap(x any) (*value, error) {
	rv := reflect.ValueOf(x)

	obj := &value{cVal: C.mj_value_new_object()}

	for _, kv := range rv.MapKeys() {
		vv := rv.MapIndex(kv)

		var k any
		if kv.Kind() == reflect.Pointer && kv.IsNil() {
			k = nil
		} else {
			k = kv.Interface()
		}

		var v any
		if vv.Kind() == reflect.Pointer && vv.IsNil() {
			v = nil
		} else {
			v = vv.Interface()
		}

		if err := obj.setKey(k, v); err != nil {
			return nil, err
		}
	}

	return obj, nil
}

func newValueNone() *value {
	return &value{cVal: C.mj_value_new_none()}
}

func newValueSeq(x any) (*value, error) {
	rv := reflect.ValueOf(x)

	l := &value{cVal: C.mj_value_new_list()}

	for i := range rv.Len() {
		val, err := newValue(rv.Index(i).Interface())
		if err != nil {
			return nil, err
		}

		if err := l.append(val); err != nil {
			return nil, err
		}
	}

	return l, nil
}

func newValueString(s string) *value {
	cStr := C.CString(s)
	defer C.free(unsafe.Pointer(cStr))

	return &value{cVal: C.mj_value_new_string(cStr)}
}

func newValueStruct(x any) (*value, error) {
	obj := &value{cVal: C.mj_value_new_object()}
	rt := reflect.TypeOf(x)
	rv := reflect.ValueOf(x)

	for i := range rv.NumField() {
		ft := rt.Field(i)
		fv := rv.Field(i)

		if !fv.CanInterface() {
			continue
		}

		name := ft.Tag.Get(tagName)
		if name == "-" {
			continue
		}
		if name == "" {
			name = ft.Name
		}

		if err := obj.setKey(name, fv.Interface()); err != nil {
			return nil, err
		}
	}

	return obj, nil
}

func newValueUint(x uint) *value {
	switch uint(math.MaxUint) {
	case math.MaxUint32:
		return newValueUint32(uint32(x))
	case math.MaxUint64:
		return newValueUint64(uint64(x))
	}

	panic(fmt.Sprintf("unexpected max uint: %d", uint(math.MaxUint)))
}

func newValueUint32(x uint32) *value {
	return &value{cVal: C.mj_value_new_u32(C.uint32_t(x))}
}

func newValueUint64(x uint64) *value {
	return &value{cVal: C.mj_value_new_u64(C.uint64_t(x))}
}
