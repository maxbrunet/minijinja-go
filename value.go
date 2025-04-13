package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #cgo noescape mj_value_append
// #cgo nocallback mj_value_append
// #cgo noescape mj_value_decref
// #cgo nocallback mj_value_decref
// #cgo noescape mj_value_get_by_index
// #cgo nocallback mj_value_get_by_index
// #cgo noescape mj_value_get_by_str
// #cgo nocallback mj_value_get_by_str
// #cgo noescape mj_value_get_by_value
// #cgo nocallback mj_value_get_by_value
// #cgo noescape mj_value_get_kind
// #cgo nocallback mj_value_get_kind
// #cgo noescape mj_value_iter_free
// #cgo nocallback mj_value_iter_free
// #cgo noescape mj_value_iter_next
// #cgo nocallback mj_value_iter_next
// #cgo noescape mj_value_len
// #cgo nocallback mj_value_len
// #cgo noescape mj_value_set_key
// #cgo nocallback mj_value_set_key
// #cgo noescape mj_value_to_str
// #cgo nocallback mj_value_to_str
// #cgo noescape mj_value_try_iter
// #cgo nocallback mj_value_try_iter
// #include <stdlib.h>
// #include <minijinja.h>
import "C"

import (
	"fmt"
	"iter"
	"math"
	"runtime"
	"unsafe"
)

const (
	tagName = "minijinja"
)

// value represents an opaque MiniJinja value.
type value struct {
	cVal C.struct_mj_value
}

// Close decrements the value refcount.
func (v *value) Close() error {
	C.mj_value_decref(&v.cVal)
	return nil
}

// setKey inserts a value into an object using a string key.
// It returns an error if the operation fails.
func (v *value) setKey(key, val any) error {
	kVal, err := newValue(key)
	if err != nil {
		return err
	}

	vVal, err := newValue(val)
	if err != nil {
		return err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if !bool(C.mj_value_set_key(&v.cVal, kVal.cVal, vVal.cVal)) {
		return getError()
	}

	return nil
}

// fieldByIndex looks up an element by an integer index in a list of object.
func (v *value) fieldByIndex(index int) *value {
	cVal := C.mj_value_get_by_index(v.cVal, C.uint64_t(index))
	return &value{cVal: cVal}
}

// fieldByName looks up an element by a string index in an object.
func (v *value) fieldByName(name string) *value {
	key := C.CString(name)
	defer C.free(unsafe.Pointer(key))

	cVal := C.mj_value_get_by_str(v.cVal, key)

	return &value{cVal: cVal}
}

// key looks up an element by a value.
func (v *value) key(key *value) *value {
	cVal := C.mj_value_get_by_value(v.cVal, key.cVal)
	return &value{cVal: cVal}
}

// append appends a value to a list.
// It returns an error if appending fails.
func (v *value) append(val *value) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if !bool(C.mj_value_append(&v.cVal, val.cVal)) {
		return getError()
	}

	return nil
}

// len returns the length of the object.
func (v *value) len() int {
	n := uint64(C.mj_value_len(v.cVal))
	if n > uint64(math.MaxInt) {
		panic(
			fmt.Sprintf(
				"value length (%d) exceeds max int (%d)",
				n,
				math.MaxInt,
			),
		)
	}

	return int(n)
}

// kind returns the kind of the value.
func (v *value) kind() valueKind {
	return valueKind(C.mj_value_get_kind(v.cVal))
}

// String converts the value to its string representation.
func (v *value) String() string {
	cStr := C.mj_value_to_str(v.cVal)
	if cStr == nil {
		panic("received nil string after converting value")
	}
	defer C.mj_str_free(cStr)

	return C.GoString(cStr)
}

// valueIter assists with iterating over a MiniJinja value.
type valueIter struct {
	ptr *C.struct_mj_value_iter
}

// newIter creates an [iter.Seq] for the value.
func (v *value) newIter() (iter.Seq[*value], error) {
	runtime.LockOSThread()
	cIter := C.mj_value_try_iter(v.cVal)
	if cIter == nil {
		return nil, getError()
	}
	runtime.UnlockOSThread()

	return func(yield func(*value) bool) {
		for {
			var cVal C.struct_mj_value
			if C.mj_value_iter_next(cIter, &cVal) != C.bool(true) {
				break
			}
			if !yield(&value{cVal: cVal}) {
				break
			}
		}
		C.mj_value_iter_free(cIter)
	}, nil
}
