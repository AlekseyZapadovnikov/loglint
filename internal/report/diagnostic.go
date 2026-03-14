package report

import (
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
	"golang.org/x/tools/go/analysis"
)

func Diagnostic(v rules.Violation) analysis.Diagnostic {
	diagnostic := analysis.Diagnostic{
		Category: string(v.Rule),
		Message:  v.Message,
	}

	if v.Expr != nil {
		diagnostic.Pos = v.Expr.Pos()
		diagnostic.End = v.Expr.End()
	}

	return diagnostic
}
