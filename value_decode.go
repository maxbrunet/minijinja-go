package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #include <stdlib.h>
// #include <minijinja.h>
import "C"

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
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
	case valueKindNumber:
		return v.decodeNumber(rv)
	case valueKindString:
		return v.decodeString(rv)
	case valueKindSeq:
		return v.decodeSeq(rv)
	case valueKindMap, valueKindIterable:
		return v.decodeMap(rv)
	case valueKindNone, valueKindUndefined:
		return nil
	case valueKindBytes, valueKindPlain, valueKindInvalid:
	}

	return &DecodeTypeError{
		Value: "unsupported " + v.kind().String(),
		Type:  rv.Type(),
	}
}

func (v *value) decodeBool(rv reflect.Value) error {
	b, err := strconv.ParseBool(v.String())
	if err != nil {
		panic(fmt.Errorf("failed to parse bool value: %w", err))
	}

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

func (v *value) decodeNone(_ reflect.Value) error {
	return nil
}

func (v *value) decodeNumber(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Float32:
		return v.decodeNumberToFloat32(rv)
	case reflect.Float64, reflect.Interface:
		return v.decodeNumberToFloat64(rv)
	case reflect.Int:
		return v.decodeNumberToInt(rv)
	case reflect.Int32:
		return v.decodeNumberToInt32(rv)
	case reflect.Int64:
		return v.decodeNumberToInt64(rv)
	case reflect.Uint:
		return v.decodeNumberToUint(rv)
	case reflect.Uint32:
		return v.decodeNumberToUint32(rv)
	case reflect.Uint64:
		return v.decodeNumberToUint64(rv)
	}

	return &DecodeTypeError{Value: v.kind().String(), Type: rv.Type()}
}

func (v *value) decodeNumberToFloat32(rv reflect.Value) error {
	s := v.String()
	n, err := strconv.ParseFloat(s, 32)
	if err != nil {
		panic(fmt.Errorf("failed to parse float32 value: %w", err))
	}

	rv.SetFloat(n)
	return nil
}

func (v *value) decodeNumberToFloat64(rv reflect.Value) error {
	s := v.String()
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse float64 value: %w", err))
	}

	switch rv.Kind() {
	case reflect.Float64:
		rv.SetFloat(n)
		return nil
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			rv.Set(reflect.ValueOf(n))
			return nil
		}
	}

	return &DecodeTypeError{Value: v.kind().String() + " " + s, Type: rv.Type()}
}

func (v *value) decodeNumberToInt(rv reflect.Value) error {
	switch math.MaxInt {
	case math.MaxInt32:
		return v.decodeNumberToInt32(rv)
	case math.MaxInt64:
		return v.decodeNumberToInt64(rv)
	}

	panic(fmt.Sprintf("unexpected max int: %d", math.MaxInt))
}

func (v *value) decodeNumberToInt32(rv reflect.Value) error {
	s := v.String()
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic(fmt.Errorf("failed to parse int32 value: %w", err))
	}

	rv.SetInt(n)
	return nil
}

func (v *value) decodeNumberToInt64(rv reflect.Value) error {
	s := v.String()
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse int64 value: %w", err))
	}

	rv.SetInt(n)
	return nil
}

func (v *value) decodeNumberToUint(rv reflect.Value) error {
	switch uint(math.MaxUint) {
	case math.MaxUint32:
		return v.decodeNumberToUint32(rv)
	case math.MaxUint64:
		return v.decodeNumberToUint64(rv)
	}

	panic(fmt.Sprintf("unexpected max uint: %d", uint(math.MaxUint)))
}

func (v *value) decodeNumberToUint32(rv reflect.Value) error {
	s := v.String()
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(fmt.Errorf("failed to parse uint32 value: %w", err))
	}

	rv.SetUint(n)
	return nil
}

func (v *value) decodeNumberToUint64(rv reflect.Value) error {
	s := v.String()
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to parse uint64 value: %w", err))
	}

	rv.SetUint(n)
	return nil
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
