package logcall

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
)

// extractMessage builds a normalized Message from an AST expression.
//
// It prefers the strongest static information first:
//  1. full compile-time constant string
//  2. structured split into literal/dynamic parts for string concatenation
//  3. fully dynamic fallback
func extractMessage(info *types.Info, expr ast.Expr) Message {
	if expr == nil {
		return Message{}
	}

	if text, ok := stringConstantValue(info, expr); ok {
		return NewStaticMessage(expr, text)
	}

	parts := splitMessageParts(info, expr)
	if len(parts) == 0 {
		return NewDynamicMessage(expr)
	}

	return NewMessageFromParts(expr, parts)
}

func splitMessageParts(info *types.Info, expr ast.Expr) []MessagePart {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.ParenExpr:
		return splitMessageParts(info, e.X)

	case *ast.BinaryExpr:
		if e.Op == token.ADD && isStringExpr(info, e) {
			left := splitMessageParts(info, e.X)
			right := splitMessageParts(info, e.Y)

			parts := make([]MessagePart, 0, len(left)+len(right))
			parts = append(parts, left...)
			parts = append(parts, right...)

			return mergeAdjacentLiteralParts(parts)
		}
	}

	if text, ok := stringConstantValue(info, expr); ok {
		return []MessagePart{
			{
				Kind: MessagePartLiteral,
				Expr: expr,
				Text: text,
			},
		}
	}

	return []MessagePart{
		{
			Kind: MessagePartDynamic,
			Expr: expr,
		},
	}
}

func mergeAdjacentLiteralParts(parts []MessagePart) []MessagePart {
	if len(parts) < 2 {
		return parts
	}

	merged := make([]MessagePart, 0, len(parts))
	for _, part := range parts {
		n := len(merged)
		if n > 0 &&
			merged[n-1].Kind == MessagePartLiteral &&
			part.Kind == MessagePartLiteral {
			merged[n-1].Text += part.Text
			continue
		}

		merged = append(merged, part)
	}

	return merged
}

func stringConstantValue(info *types.Info, expr ast.Expr) (string, bool) {
	if info == nil || expr == nil {
		return "", false
	}

	tv, ok := info.Types[expr]
	if !ok || tv.Value == nil || tv.Type == nil {
		return "", false
	}

	if !isStringType(tv.Type) {
		return "", false
	}

	return constant.StringVal(tv.Value), true
}

func isStringExpr(info *types.Info, expr ast.Expr) bool {
	if info == nil || expr == nil {
		return false
	}

	tv, ok := info.Types[expr]
	if !ok || tv.Type == nil {
		return false
	}

	return isStringType(tv.Type)
}

func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}

	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	return basic.Info()&types.IsString != 0
}
