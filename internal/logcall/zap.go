package logcall

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const (
	zapPackagePath       = "go.uber.org/zap"
	zapLoggerName        = "Logger"
	zapSugaredLoggerName = "SugaredLogger"
	zapFieldTypeName     = "Field"
)

func extractZap(pass *analysis.Pass, call *ast.CallExpr) (Record, bool, error) {
	record, matched, err := extractZapLogger(pass, call)
	if matched {
		return record, true, err
	}

	record, matched, err = extractZapSugared(pass, call)
	if matched {
		return record, true, err
	}

	return Record{}, false, nil
}

func extractZapLogger(pass *analysis.Pass, call *ast.CallExpr) (Record, bool, error) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Record{}, false, nil
	}

	method := sel.Sel.Name
	level, ok := normalizeZapLoggerLevel(method)
	if !ok {
		return Record{}, false, nil
	}

	if !isMethodOfNamedType(pass.TypesInfo, sel, method, zapPackagePath, zapLoggerName) {
		return Record{}, false, nil
	}

	if len(call.Args) == 0 {
		return Record{}, true, ErrMissingMessage
	}

	return Record{
		Kind:     LoggerZap,
		Level:    level,
		Method:   method,
		Call:     call,
		Selector: sel,
		Message:  extractMessage(pass.TypesInfo, call.Args[0]),
		Fields:   extractZapLoggerFields(pass.TypesInfo, call.Args[1:]),
	}, true, nil
}

func extractZapSugared(pass *analysis.Pass, call *ast.CallExpr) (Record, bool, error) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return Record{}, false, nil
	}

	method := sel.Sel.Name
	level, ok := normalizeZapSugaredLevel(method)
	if !ok {
		return Record{}, false, nil
	}

	if !isMethodOfNamedType(pass.TypesInfo, sel, method, zapPackagePath, zapSugaredLoggerName) {
		return Record{}, false, nil
	}

	if len(call.Args) == 0 {
		return Record{}, true, ErrMissingMessage
	}

	return Record{
		Kind:     LoggerZapSugared,
		Level:    level,
		Method:   method,
		Call:     call,
		Selector: sel,
		Message:  extractMessage(pass.TypesInfo, call.Args[0]),
		Fields:   extractZapSugaredFields(pass.TypesInfo, call.Args[1:]),
	}, true, nil
}

func normalizeZapLoggerLevel(method string) (Level, bool) {
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

func normalizeZapSugaredLevel(method string) (Level, bool) {
	switch method {
	case "Debugw":
		return LevelDebug, true
	case "Infow":
		return LevelInfo, true
	case "Warnw":
		return LevelWarn, true
	case "Errorw":
		return LevelError, true
	default:
		return LevelUnknown, false
	}
}

func isMethodOfNamedType(
	info *types.Info,
	sel *ast.SelectorExpr,
	method string,
	pkgPath string,
	typeName string,
) bool {
	if info == nil || sel == nil || sel.Sel == nil {
		return false
	}

	selection := info.Selections[sel]
	if selection == nil {
		return false
	}

	fn, ok := selection.Obj().(*types.Func)
	if !ok || fn.Name() != method {
		return false
	}

	if fn.Pkg() == nil || !matchesPackagePath(fn.Pkg().Path(), pkgPath) {
		return false
	}

	return isNamedType(selection.Recv(), pkgPath, typeName)
}

func extractZapLoggerFields(info *types.Info, args []ast.Expr) []Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]Field, 0, len(args))
	for _, arg := range args {
		field, ok := extractZapField(info, arg)
		if !ok {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

func extractZapSugaredFields(info *types.Info, args []ast.Expr) []Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]Field, 0, len(args))

	for i := 0; i < len(args); {
		if field, ok := extractZapField(info, args[i]); ok {
			fields = append(fields, field)
			i++
			continue
		}

		if i+1 >= len(args) {
			break
		}

		keyExpr := args[i]
		valueExpr := args[i+1]
		key, ok := stringConstantValue(info, keyExpr)

		fields = append(fields, Field{
			Key:       key,
			KeyExpr:   keyExpr,
			ValueExpr: valueExpr,
			StaticKey: ok,
		})

		i += 2
	}

	return fields
}

func extractZapField(info *types.Info, expr ast.Expr) (Field, bool) {
	if !isZapFieldExpr(info, expr) {
		return Field{}, false
	}

	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return Field{
			ValueExpr: expr,
			StaticKey: false,
		}, true
	}

	keyExpr := call.Args[0]
	key, ok := stringConstantValue(info, keyExpr)

	return Field{
		Key:       key,
		KeyExpr:   keyExpr,
		ValueExpr: expr,
		StaticKey: ok,
	}, true
}

func isZapFieldExpr(info *types.Info, expr ast.Expr) bool {
	if info == nil || expr == nil {
		return false
	}

	tv, ok := info.Types[expr]
	if !ok || tv.Type == nil {
		return false
	}

	return isNamedType(tv.Type, zapPackagePath, zapFieldTypeName)
}
