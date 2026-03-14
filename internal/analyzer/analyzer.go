package analyzer

import (
	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const (
	Name = "loglint"
	Doc  = "checks log calls and log messages"
)

// Analyzer is the default analyzer instance with default configuration.
// It is kept for backward compatibility and standalone usage.
var Analyzer = MustNew(config.DefaultConfig())

// MustNew creates a new analyzer with the given configuration.
// It panics if the configuration is invalid.
// Use New for error handling instead.
func MustNew(cfg config.Config) *analysis.Analyzer {
	a, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return a
}

// New creates a new analyzer with the given configuration.
// Returns an error if the configuration is invalid.
func New(cfg config.Config) (*analysis.Analyzer, error) {
	cfg = cfg.Normalize()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ruleSet, err := rules.NewRuleSet(cfg)
	if err != nil {
		return nil, err
	}

	return &analysis.Analyzer{
		Name:     Name,
		Doc:      Doc,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      newRun(ruleSet),
	}, nil
}
