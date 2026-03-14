package ruleid

import "strings"

type ID string

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

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func OrderedRules() []ID {
	return append([]ID(nil), orderedRules...)
}

func AllowedEnabledNames() []string {
	return []string{
		string(All),
		string(Lowercase),
		string(English),
		string(Symbols),
		string(Sensitive),
	}
}

func AllowedRuleNames() []string {
	return []string{
		string(Lowercase),
		string(English),
		string(Symbols),
		string(Sensitive),
	}
}

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
