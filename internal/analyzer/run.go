package analyzer

import (
	"errors"
	"fmt"
	"go/ast"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/report"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
	"golang.org/x/tools/go/analysis"
	inspectpass "golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const extractionErrorReportLimit = 5

// newRun returns a Run function for the analyzer with the given configuration.
func newRun(ruleSet *rules.RuleSet) func(pass *analysis.Pass) (any, error) {
	return func(pass *analysis.Pass) (any, error) {
		insp := pass.ResultOf[inspectpass.Analyzer].(*inspector.Inspector)
		reportedExtractionErrors := make(map[string]struct{})

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

				// Do not silently swallow extraction failures:
				// report each distinct error up to a small limit to avoid flooding output.
				key := fmt.Sprintf("%T:%v", err, err)
				if _, exists := reportedExtractionErrors[key]; !exists &&
					len(reportedExtractionErrors) < extractionErrorReportLimit {
					pass.Reportf(call.Pos(), "loglint: failed to analyze log call: %v", err)
					reportedExtractionErrors[key] = struct{}{}
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
