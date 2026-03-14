package loglint

import (
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(
		t,
		testdata,
		analyzer.Analyzer,
		"slog/english",
		"slog/ignored",
		"slog/lowercase",
		"slog/mixed",
		"slog/ok",
		"slog/sensitive_field",
		"slog/sensitive_message",
		"slog/symbols",
		"zap/logger",
		"zap/ok",
		"zap/sugared",
		"zap/unsupported",
	)
}
