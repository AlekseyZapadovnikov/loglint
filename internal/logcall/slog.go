package logcall

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const (
	slogPackagePath = "log/slog"
	slogLoggerName  = "Logger"
)

func extractSlog(pass *analysis.Pass, call *ast.CallExpr) (Record, bool, error) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Record{}, false, nil
	}

	method := sel.Sel.Name
	level, ok := normalizeSlogLevel(method)
	if !ok {
		return Record{}, false, nil
	}

	if !isSupportedSlogCall(pass.TypesInfo, sel, method) {
		return Record{}, false, nil
	}

	if len(call.Args) == 0 {
		return Record{}, true, ErrMissingMessage
	}

	record := Record{
		Kind:     LoggerSlog,
		Level:    level,
		Method:   method,
		Call:     call,
		Selector: sel,
		Message:  extractMessage(pass.TypesInfo, call.Args[0]),
		Fields:   extractSlogFields(pass.TypesInfo, call.Args[1:]),
	}

	return record, true, nil
}

func normalizeSlogLevel(method string) (Level, bool) {
	switch method {
	case "Debug":
		return LevelDebug, true
	case "Info":
		return LevelInfo, true
	case "Warn":
		return LevelWarn, true
	case "Error":
		return LevelError, true
	default:
		return LevelUnknown, false
	}
}

func isSupportedSlogCall(info *types.Info, sel *ast.SelectorExpr, method string) bool {
	return isSlogPackageFunc(info, sel.X) || isSlogLoggerMethod(info, sel, method)
}

func isSlogPackageFunc(info *types.Info, x ast.Expr) bool {
	pkgIdent, ok := x.(*ast.Ident)
	if !ok {
		return false
	}

	pkgName, ok := info.Uses[pkgIdent].(*types.PkgName)
	if !ok || pkgName.Imported() == nil {
		return false
	}

	return matchesPackagePath(pkgName.Imported().Path(), slogPackagePath)
}

func isSlogLoggerMethod(info *types.Info, sel *ast.SelectorExpr, method string) bool {
	selection := info.Selections[sel]
	if selection == nil {
		return false
	}

	fn, ok := selection.Obj().(*types.Func)
	if !ok || fn.Name() != method {
		return false
	}

	if fn.Pkg() == nil || !matchesPackagePath(fn.Pkg().Path(), slogPackagePath) {
		return false
	}

	return isNamedType(selection.Recv(), slogPackagePath, slogLoggerName)
}

func isNamedType(t types.Type, pkgPath, typeName string) bool {
	if t == nil {
		return false
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	return obj != nil &&
		obj.Pkg() != nil &&
		matchesPackagePath(obj.Pkg().Path(), pkgPath) &&
		obj.Name() == typeName
}

// extractSlogFields currently supports only alternating key-value pairs
// with compile-time string keys:
//
//	slog.Info("msg", "user_id", id, "request_id", reqID)
//
// Unsupported forms such as slog.Attr are intentionally skipped for now.
func extractSlogFields(info *types.Info, args []ast.Expr) []Field {
	if len(args) < 2 {
		return nil
	}

	fields := make([]Field, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		keyExpr := args[i]
		valueExpr := args[i+1]

		key, ok := stringConstantValue(info, keyExpr)

		fields = append(fields, Field{
			Key:       key,
			KeyExpr:   keyExpr,
			ValueExpr: valueExpr,
			StaticKey: ok,
		})
	}

	return fields
}
