// Package minijinja provides Go bindings for the [MiniJinja] Rust library.
//
// [MiniJinja]: https://docs.rs/minijinja/
package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #include <stdlib.h>
// #include <minijinja.h>
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"
)

// Environment represents a MiniJinja environment.
type Environment struct {
	ptr *C.struct_mj_env
}

// NewEnvironment allocates and returns a new, empty MiniJinja environment.
func NewEnvironment() *Environment {
	env := C.mj_env_new()
	if env == nil {
		panic("received nil env")
	}

	return &Environment{ptr: env}
}

// Close closes the environment.
func (e *Environment) Close() error {
	if e.ptr != nil {
		C.mj_env_free(e.ptr)
		e.ptr = nil
	}
	return nil
}

// SetKeepTrailingNewline preserves the trailing newline when rendering
// templates.
func (e *Environment) SetKeepTrailingNewline(on bool) {
	C.mj_env_set_keep_trailing_newline(e.ptr, C.bool(on))
}

// SetLStripBlocks enables or disables the lstrip_blocks feature.
func (e *Environment) SetLStripBlocks(on bool) {
	C.mj_env_set_lstrip_blocks(e.ptr, C.bool(on))
}

// SetRecursionLimit changes the recursion limit.
func (e *Environment) SetRecursionLimit(limit uint) {
	C.mj_env_set_recursion_limit(e.ptr, C.uint32_t(limit))
}

// SetSyntaxConfig reconfigures the syntax.
func (e *Environment) SetSyntaxConfig(syntax *SyntaxConfig) error {
	cStx := newCSyntaxConfig(syntax)
	defer cStx.Close()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if !C.mj_env_set_syntax_config(e.ptr, cStx.ptr) {
		return getError()
	}

	return nil
}

// SetTrimBlocks enables or disables the trim_blocks feature.
func (e *Environment) SetTrimBlocks(on bool) {
	C.mj_env_set_trim_blocks(e.ptr, C.bool(on))
}

// UndefinedBehavior controls the undefined behavior of the engine.
type UndefinedBehavior int

const (
	// UndefinedBehaviorLenient is the default, somewhat lenient undefined
	// behavior.
	UndefinedBehaviorLenient UndefinedBehavior = iota
	// UndefinedBehaviorStrict complains very quickly about undefined values.
	UndefinedBehaviorStrict
	// UndefinedBehaviorChainable is like lenient, but also allows chaining of
	// undefined lookups.
	UndefinedBehaviorChainable
)

// SetUndefinedBehavior reconfigures the undefined behavior.
func (e *Environment) SetUndefinedBehavior(behavior UndefinedBehavior) {
	C.mj_env_set_undefined_behavior(e.ptr, C.enum_mj_undefined_behavior(behavior))
}

// AddTemplate registers a template with the environment.
func (e *Environment) AddTemplate(name, source string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if ok := C.mj_env_add_template(e.ptr, cName, cSource); !ok {
		return getError()
	}

	return nil
}

// RemoveTemplate removes a template from the environment.
func (e *Environment) RemoveTemplate(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if ok := C.mj_env_remove_template(e.ptr, cName); !ok {
		return getError()
	}

	return nil
}

// ClearTemplates clears all templates.
func (e *Environment) ClearTemplates() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if ok := C.mj_env_clear_templates(e.ptr); !ok {
		return getError()
	}

	return nil
}

// RenderNamedString renders a template from a named string.
func (e *Environment) RenderNamedString(
	name, source string, ctx any,
) (string, error) {
	val, err := newValue(ctx)
	if err != nil {
		return "", err
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cSrc := C.CString(source)
	defer C.free(unsafe.Pointer(cSrc))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	out := C.mj_env_render_named_str(e.ptr, cName, cSrc, val.cVal)
	if out == nil {
		return "", getError()
	}
	defer C.mj_str_free(out)

	return C.GoString(out), nil
}

// RenderTemplate renders a registered template using the provided context.
func (e *Environment) RenderTemplate(name string, ctx any) (string, error) {
	val, err := newValue(ctx)
	if err != nil {
		return "", err
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	out := C.mj_env_render_template(e.ptr, cName, val.cVal)
	if out == nil {
		return "", getError()
	}
	defer C.mj_str_free(out)

	return C.GoString(out), nil
}

// EvalExpr evaluates an expression string in the given context.
// It stores the result in the value pointed by data or returns an error if the
// evaluation fails.
func (e *Environment) EvalExpr(expr string, ctx, data any) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidEvalExprError{Type: reflect.TypeOf(data)}
	}

	val, err := newValue(ctx)
	if err != nil {
		return err
	}

	cExpr := C.CString(expr)
	defer C.free(unsafe.Pointer(cExpr))

	runtime.LockOSThread()
	cRes := C.mj_env_eval_expr(e.ptr, cExpr, val.cVal)
	if isErrorSet() {
		return getError()
	}
	runtime.UnlockOSThread()

	res := &value{cVal: cRes}
	defer res.Close()
	return res.decode(rv.Elem())
}
