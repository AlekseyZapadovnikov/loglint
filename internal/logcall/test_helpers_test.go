package logcall

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

type checkedSource struct {
	file *ast.File
	info *types.Info
}

func mustCheckSource(t *testing.T, src string) checkedSource {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	cfg := &types.Config{
		Importer: importer.Default(),
	}

	if _, err := cfg.Check("p", fset, []*ast.File{file}, info); err != nil {
		t.Fatalf("type-check source: %v", err)
	}

	return checkedSource{
		file: file,
		info: info,
	}
}

func firstAssignRHS(t *testing.T, file *ast.File) ast.Expr {
	t.Helper()

	var expr ast.Expr
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) == 0 {
			return true
		}

		expr = assign.Rhs[0]
		return false
	})

	if expr == nil {
		t.Fatal("assignment RHS not found")
	}

	return expr
}

func callExpressions(file *ast.File) []*ast.CallExpr {
	calls := make([]*ast.CallExpr, 0, 8)
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok {
			calls = append(calls, call)
		}
		return true
	})

	return calls
}

func callsByMethod(file *ast.File, method string) []*ast.CallExpr {
	calls := make([]*ast.CallExpr, 0, 4)

	for _, call := range callExpressions(file) {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel == nil || sel.Sel.Name != method {
			continue
		}

		calls = append(calls, call)
	}

	return calls
}
