package rules

import (
	"go/ast"

	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/ruleid"
)

type ID = ruleid.ID

const (
	RuleLowercase ID = ruleid.Lowercase
	RuleEnglish   ID = ruleid.English
	RuleSymbols   ID = ruleid.Symbols
	RuleSensitive ID = ruleid.Sensitive
)

type Violation struct {
	Rule    ID
	Message string
	Expr    ast.Expr
}

type CheckFunc func(logcall.Record) []Violation

// RuleSet contains all rule checkers.
// It supports configuration for rules that need it (e.g., sensitive).
type RuleSet struct {
	checks []CheckFunc
}

// NewRuleSet creates a new RuleSet with the given configuration.
func NewRuleSet(cfg config.Config) (*RuleSet, error) {
	resolved, err := cfg.ResolveRules()
	if err != nil {
		return nil, err
	}

	checks := make([]CheckFunc, 0, len(resolved))
	sensitive := NewSensitiveChecker(cfg.Sensitive)

	for _, id := range resolved {
		switch id {
		case ruleid.Lowercase:
			checks = append(checks, CheckLowercase)
		case ruleid.English:
			checks = append(checks, CheckEnglish)
		case ruleid.Symbols:
			checks = append(checks, CheckSymbols)
		case ruleid.Sensitive:
			checks = append(checks, sensitive.Check)
		}
	}

	return &RuleSet{checks: checks}, nil
}

// Apply applies all rules to the given log record.
func (rs *RuleSet) Apply(record logcall.Record) []Violation {
	var violations []Violation

	for _, check := range rs.checks {
		violations = append(violations, check(record)...)
	}

	return violations
}

// defaultRuleSet is the default rule set with default configuration.
// Used for backward compatibility with Apply function.
var defaultRuleSet = mustNewRuleSet(config.DefaultConfig())

// Apply applies all rules with default configuration to the given log record.
// This function is kept for backward compatibility.
// For custom configuration, use NewRuleSet and RuleSet.Apply instead.
func Apply(record logcall.Record) []Violation {
	return defaultRuleSet.Apply(record)
}

func mustNewRuleSet(cfg config.Config) *RuleSet {
	rs, err := NewRuleSet(cfg)
	if err != nil {
		panic(err)
	}
	return rs
}

func newViolation(rule ID, expr ast.Expr, message string) Violation {
	return Violation{
		Rule:    rule,
		Message: message,
		Expr:    expr,
	}
}

func staticMessageFragments(msg logcall.Message) []string {
	if len(msg.Parts) == 0 {
		return nil
	}

	fragments := make([]string, 0, len(msg.Parts))
	for _, part := range msg.Parts {
		if !part.IsStatic() || part.Text == "" {
			continue
		}

		fragments = append(fragments, part.Text)
	}

	return fragments
}
