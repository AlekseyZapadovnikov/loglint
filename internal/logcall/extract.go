package logcall

import (
	"errors"
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var (
	ErrNotLogCall     = errors.New("not a supported log call")
	ErrMissingMessage = errors.New("log call has no message argument")
)

// Extract converts a raw call expression into a normalized Record.
func Extract(pass *analysis.Pass, call *ast.CallExpr) (Record, error) {
	if pass == nil {
		return Record{}, fmt.Errorf("extract log call: nil analysis pass")
	}

	if call == nil {
		return Record{}, fmt.Errorf("extract log call: nil call expression")
	}

	if record, matched, err := extractSlog(pass, call); matched {
		return record, err
	}

	if record, matched, err := extractZap(pass, call); matched {
		return record, err
	}

	return Record{}, ErrNotLogCall
}
