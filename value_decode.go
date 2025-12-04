package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #cgo noescape mj_value_as_f64
// #cgo nocallback mj_value_as_f64
// #cgo noescape mj_value_as_i64
// #cgo nocallback mj_value_as_i64
// #cgo noescape mj_value_as_u64
// #cgo nocallback mj_value_as_u64
// #cgo noescape mj_value_is_true
// #cgo nocallback mj_value_is_true
// #include <stdlib.h>
// #include <minijinja.h>
import "C"

import (
	"encoding"
	"fmt"
	"reflect"
	"unsafe"
)

var (
	binaryUnmarshalerType = reflect.TypeFor[encoding.BinaryUnmarshaler]()
	textUnmarshalerType   = reflect.TypeFor[encoding.TextUnmarshaler]()
)

// decode decodes a value and stores the result into the variable pointed by rv.
func (v *value) decode(rv reflect.Value) error {
	kind := v.kind()
	if rv.Kind() == reflect.Pointer {
		if kind == valueKindNone || kind == valueKindUndefined {
			return nil
		}

		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return v.decode(rv.Elem())
	}

	switch kind {
	case valueKindBool:
		return v.decodeBool(rv)
	case valueKindBytes:
		return v.decodeBytes(rv)
	case valueKindNumber:
		return v.decodeNumber(rv)
	case valueKindPlain, valueKindString:
		return v.decodeString(rv)
	case valueKindSeq:
		return v.decodeSeq(rv)
	case valueKindMap, valueKindIterable:
		return v.decodeMap(rv)
	case valueKindNone, valueKindUndefined:
		return nil
	case valueKindInvalid:
	}

	return &DecodeTypeError{
		Value: "unsupported " + v.kind().String(), Type: rv.Type(),
	}
}

func (v *value) decodeBool(rv reflect.Value) error {
	b := bool(C.mj_value_is_true(v.cVal))

	switch rv.Kind() {
	case reflect.Bool:
		rv.SetBool(b)
		return nil
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			rv.Set(reflect.ValueOf(b))
			return nil
		}
	}

	return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
}

func (v *value) decodeBytes(rv reflect.Value) error {
	length := C.uintptr_t(0)
	charPtr := C.mj_value_as_bytes(v.cVal, &length)
	bytes := C.GoBytes(unsafe.Pointer(charPtr), C.int(length))

	rt := rv.Type()
	if reflect.PointerTo(rt).Implements(binaryUnmarshalerType) {
		uv := reflect.New(rt)
		if u, ok := uv.Interface().(encoding.BinaryUnmarshaler); ok {
			err := u.UnmarshalBinary(bytes)
			if err != nil {
				return &UnmarshalerError{
					Type:       rt,
					Err:        err,
					sourceFunc: "UnmarshalBinary",
				}
			}
			rv.Set(uv.Elem())
			return nil
		}
	}

	switch rv.Kind() {
	case reflect.Array:
		return v.decodeBytesToArray(rv, bytes)
	case reflect.Slice:
		return v.decodeBytesToSlice(rv, bytes)
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			return v.decodeBytesToSlice(rv, bytes)
		}
	}

	return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
}

func (v *value) decodeBytesToArray(rv reflect.Value, bytes []byte) error {
	rt := rv.Type().Elem().Kind()
	for i, b := range bytes {
		if i >= rv.Len() {
			// Ran out of fixed array: skip.
			return nil
		}

		switch rt {
		case reflect.Uint8:
			rv.Index(i).SetUint(uint64(b))
		case reflect.Interface:
			e := rv.Index(i)
			if e.NumMethod() == 0 {
				e.Set(reflect.ValueOf(b))
				continue
			}
			fallthrough
		default:
			return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
		}
	}

	if len(bytes) < rv.Len() {
		for i := len(bytes); i < rv.Len(); i++ {
			// zero remainder of array
			rv.Index(i).SetZero()
		}
	}

	return nil
}

func (v *value) decodeBytesToSlice(rv reflect.Value, bytes []byte) error {
	var s reflect.Value
	if rv.Kind() == reflect.Interface {
		if rv.NumMethod() > 0 {
			return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
		}
		s = reflect.New(reflect.TypeFor[[]byte]()).Elem()
	} else {
		if rv.IsNil() && rv.Type().Elem().Kind() == reflect.Uint8 {
			rv.SetBytes(bytes)
			return nil
		}
		s = rv
	}

	if s.Cap() < len(bytes) {
		s.Grow(len(bytes) - s.Cap())
	}
	s.SetLen(len(bytes))

	st := s.Type().Elem().Kind()
	for i, b := range bytes {
		switch st {
		case reflect.Uint8:
			s.Index(i).SetUint(uint64(b))
		case reflect.Interface:
			e := s.Index(i)
			if e.NumMethod() == 0 {
				e.Set(reflect.ValueOf(b))
				continue
			}
			fallthrough
		default:
			return &DecodeTypeError{Value: v.kind().String(), Type: s.Type()}
		}
	}

	rv.Set(s)
	return nil
}

func (v *value) decodeMap(rv reflect.Value) error {
	var rt reflect.Type
	switch rv.Kind() {
	case reflect.Interface:
		if rv.NumMethod() > 0 {
			return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
		}
		rt = reflect.MapOf(rv.Type(), rv.Type())
		rv.Set(reflect.MakeMapWithSize(rt, v.len()))
		rv = rv.Elem()
	case reflect.Map:
		rt = rv.Type()
		if rv.IsNil() {
			rv.Set(reflect.MakeMapWithSize(rt, v.len()))
		}
	case reflect.Struct:
		return v.decodeMapToStruct(rv)
	default:
		return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
	}

	vIter, err := v.newIter()
	if err != nil {
		return err
	}

	keyType := rt.Key()
	valueType := rt.Elem()

	for key := range vIter {
		val := v.key(key)

		kv := reflect.New(keyType)
		if err := key.decode(kv); err != nil {
			return err
		}

		vv := reflect.New(valueType)
		if err := val.decode(vv); err != nil {
			return err
		}

		rv.SetMapIndex(kv.Elem(), vv.Elem())
	}

	return nil
}

func (v *value) decodeMapToStruct(rv reflect.Value) error {
	rt := rv.Type()

	for i := range rt.NumField() {
		ft := rt.Field(i)
		fv := rv.Field(i)

		if !fv.CanSet() {
			continue
		}

		name := ft.Tag.Get(tagName)
		if name == "-" {
			continue
		}
		if name == "" {
			name = ft.Name
		}

		val := v.fieldByName(name)

		if err := val.decode(fv); err != nil {
			return err
		}
	}

	return nil
}

func (v *value) decodeNumber(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		rv.SetFloat(float64(C.mj_value_as_f64(v.cVal)))
		return nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		rv.SetInt(int64(C.mj_value_as_i64(v.cVal)))
		return nil
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			rv.Set(reflect.ValueOf(float64(C.mj_value_as_f64(v.cVal))))
			return nil
		}
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		rv.SetUint(uint64(C.mj_value_as_u64(v.cVal)))
		return nil
	}

	return &DecodeTypeError{
		Value: fmt.Sprintf("%s %s", v.kind(), v), Type: rv.Type(),
	}
}

func (v *value) decodeSeq(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Array:
		return v.decodeSeqToArray(rv)
	case reflect.Interface, reflect.Slice:
		return v.decodeSeqToSlice(rv)
	}

	return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
}

func (v *value) decodeSeqToArray(rv reflect.Value) error {
	for i := range v.len() {
		if i >= rv.Len() {
			// Ran out of fixed array: skip.
			return nil
		}

		val := v.fieldByIndex(i)

		if err := val.decode(rv.Index(i)); err != nil {
			return err
		}
	}

	if v.len() < rv.Len() {
		for i := v.len(); i < rv.Len(); i++ {
			// zero remainder of array
			rv.Index(i).SetZero()
		}
	}

	return nil
}

func (v *value) decodeSeqToSlice(rv reflect.Value) error {
	var s reflect.Value
	if rv.Kind() == reflect.Interface {
		if rv.NumMethod() > 0 {
			return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
		}
		rt := reflect.SliceOf(rv.Type())
		s = reflect.New(rt).Elem()
	} else {
		s = rv
	}

	if s.Cap() < v.len() {
		s.Grow(v.len() - s.Cap())
	}
	s.SetLen(v.len())

	for i := range v.len() {
		val := v.fieldByIndex(i)

		if err := val.decode(s.Index(i)); err != nil {
			return err
		}
	}

	rv.Set(s)
	return nil
}

func (v *value) decodeString(rv reflect.Value) error {
	s := v.String()

	rt := rv.Type()
	if reflect.PointerTo(rt).Implements(textUnmarshalerType) {
		uv := reflect.New(rt)
		if u, ok := uv.Interface().(encoding.TextUnmarshaler); ok {
			err := u.UnmarshalText([]byte(s))
			if err != nil {
				return &UnmarshalerError{
					Type:       rt,
					Err:        err,
					sourceFunc: "UnmarshalText",
				}
			}
			rv.Set(uv.Elem())
			return nil
		}
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(s)
		return nil
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			rv.Set(reflect.ValueOf(s))
			return nil
		}
	}

	return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
}
