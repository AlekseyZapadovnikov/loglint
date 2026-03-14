package analyzer

import (
	"errors"
	"go/ast"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/report"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
	"golang.org/x/tools/go/analysis"
	inspectpass "golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// newRun returns a Run function for the analyzer with the given configuration.
func newRun(ruleSet *rules.RuleSet) func(pass *analysis.Pass) (any, error) {
	return func(pass *analysis.Pass) (any, error) {
		insp := pass.ResultOf[inspectpass.Analyzer].(*inspector.Inspector)

		nodeFilter := []ast.Node{
			(*ast.CallExpr)(nil),
		}

		insp.Preorder(nodeFilter, func(n ast.Node) {
			call := n.(*ast.CallExpr)

			record, err := logcall.Extract(pass, call)
			if err != nil {
				if errors.Is(err, logcall.ErrNotLogCall) {
					return
				}

				return
			}

			for _, violation := range ruleSet.Apply(record) {
				pass.Report(report.Diagnostic(violation))
			}
		})

		return nil, nil
	}
}
