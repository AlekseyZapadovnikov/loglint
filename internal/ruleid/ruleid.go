package ruleid

import "strings"

// ID is a canonical rule identifier.
type ID string

// Supported rule identifiers.
const (
	All       ID = "all"
	Lowercase ID = "lowercase"
	English   ID = "english"
	Symbols   ID = "symbols"
	Sensitive ID = "sensitive"
)

var orderedRules = []ID{
	Lowercase,
	English,
	Symbols,
	Sensitive,
}

// NormalizeName trims spaces and lowercases a rule name.
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// OrderedRules returns deterministic rule order used by config resolution.
func OrderedRules() []ID {
	return append([]ID(nil), orderedRules...)
}

// AllowedEnabledNames returns allowed values for enabled_rules.
func AllowedEnabledNames() []string {
	return []string{
		string(All),
		string(Lowercase),
		string(English),
		string(Symbols),
		string(Sensitive),
	}
}

// AllowedRuleNames returns allowed values for disabled_rules.
func AllowedRuleNames() []string {
	return []string{
		string(Lowercase),
		string(English),
		string(Symbols),
		string(Sensitive),
	}
}

// ParseRule converts a user-provided rule name into ID.
func ParseRule(name string) (ID, bool) {
	switch NormalizeName(name) {
	case string(Lowercase):
		return Lowercase, true
	case string(English):
		return English, true
	case string(Symbols):
		return Symbols, true
	case string(Sensitive):
		return Sensitive, true
	default:
		return "", false
	}
}
