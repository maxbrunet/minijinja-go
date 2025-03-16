package minijinja

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -L${SRCDIR}/lib -lminijinja_cabi
// #cgo noescape mj_syntax_config_default
// #cgo nocallback mj_syntax_config_default
// #include <stdlib.h>
// #include <minijinja.h>
import "C"
import "unsafe"

// SyntaxConfig allows one to override the syntax elements.
type SyntaxConfig struct {
	BlockStart          string // Block start delimiter.
	BlockEnd            string // Block end delimiter.
	VariableStart       string // Variable start delimiter.
	VariableEnd         string // Variable end delimiter.
	CommentStart        string // Comment start delimiter.
	CommentEnd          string // Comment end delimiter.
	LineStatementPrefix string // Line statement prefix.
	LineCommentPrefix   string // Line comment prefix.
}

type cSyntaxConfig struct {
	ptr *C.mj_syntax_config
}

func newCSyntaxConfig(syntax *SyntaxConfig) *cSyntaxConfig {
	return &cSyntaxConfig{
		ptr: &C.struct_mj_syntax_config{
			block_start:           C.CString(syntax.BlockStart),
			block_end:             C.CString(syntax.BlockEnd),
			variable_start:        C.CString(syntax.VariableStart),
			variable_end:          C.CString(syntax.VariableEnd),
			comment_start:         C.CString(syntax.CommentStart),
			comment_end:           C.CString(syntax.CommentEnd),
			line_statement_prefix: C.CString(syntax.LineStatementPrefix),
			line_comment_prefix:   C.CString(syntax.LineCommentPrefix),
		},
	}
}

func (s *cSyntaxConfig) Close() error {
	if s.ptr == nil {
		return nil
	}

	C.free(unsafe.Pointer(s.ptr.block_start))
	C.free(unsafe.Pointer(s.ptr.block_end))
	C.free(unsafe.Pointer(s.ptr.variable_start))
	C.free(unsafe.Pointer(s.ptr.variable_end))
	C.free(unsafe.Pointer(s.ptr.comment_start))
	C.free(unsafe.Pointer(s.ptr.comment_end))
	C.free(unsafe.Pointer(s.ptr.line_statement_prefix))
	C.free(unsafe.Pointer(s.ptr.line_comment_prefix))

	s.ptr = nil

	return nil
}

// SyntaxConfigDefaults sets the syntax to defaults.
func SyntaxConfigDefaults(syntax *SyntaxConfig) {
	cStx := newCSyntaxConfig(syntax)
	defer cStx.Close()
	C.mj_syntax_config_default(cStx.ptr)
}
