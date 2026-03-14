package rules

import (
	"strings"
	"unicode"

	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
)

// SensitiveChecker checks for sensitive data in log messages and fields.
// It supports custom keywords via configuration.
type SensitiveChecker struct {
	// keywords is a slice of normalized keyword sequences.
	// Each keyword is stored as a slice of words (e.g., "api_key" -> ["api", "key"]).
	keywords [][]string
}

// NewSensitiveChecker creates a new SensitiveChecker with the given configuration.
// If cfg.ReplaceDefaults is true, only ExtraKeywords are used.
// Otherwise, ExtraKeywords are added to the default set.
func NewSensitiveChecker(cfg config.SensitiveConfig) *SensitiveChecker {
	cfg = cfg.Normalize()

	var keywords [][]string

	if cfg.ReplaceDefaults {
		keywords = buildKeywordsFromCanonical(cfg.ExtraKeywords)
	} else {
		allKeywords := make([]string, 0, len(config.DefaultSensitiveKeywords)+len(cfg.ExtraKeywords))
		allKeywords = append(allKeywords, config.DefaultSensitiveKeywords...)
		allKeywords = append(allKeywords, cfg.ExtraKeywords...)
		keywords = buildKeywordsFromCanonical(allKeywords)
	}

	return &SensitiveChecker{
		keywords: keywords,
	}
}

// buildKeywordsFromCanonical converts canonical keywords (e.g., "api_key") to word sequences.
// Keywords are already in canonical form (lowercase, underscore-separated).
func buildKeywordsFromCanonical(keywords []string) [][]string {
	result := make([][]string, 0, len(keywords))
	for _, kw := range keywords {
		seq := strings.Split(kw, "_")
		filtered := make([]string, 0, len(seq))
		for _, s := range seq {
			if s != "" {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) > 0 {
			result = append(result, filtered)
		}
	}
	return result
}

// Check checks a log record for sensitive data violations.
func (c *SensitiveChecker) Check(record logcall.Record) []Violation {
	if !record.HasMessage() && len(record.Fields) == 0 {
		return nil
	}

	if c.hasSensitiveMessage(record) {
		return []Violation{
			newViolation(
				RuleSensitive,
				record.Message.Expr,
				"log message may contain sensitive data",
			),
		}
	}

	for _, field := range record.Fields {
		if !field.HasStaticKey() {
			continue
		}

		if !c.containsSensitiveToken(field.Key) {
			continue
		}

		expr := field.KeyExpr
		if expr == nil {
			expr = field.ValueExpr
		}

		return []Violation{
			newViolation(
				RuleSensitive,
				expr,
				"structured log field may contain sensitive data",
			),
		}
	}

	return nil
}

// hasSensitiveMessage checks if the message may contain sensitive data.
// Message semantics:
//  1. Static fragments are checked for assignment-like patterns (":" or "=")
//     near sensitive keywords (e.g. "password: value", "token=abc").
//  2. Dynamic messages additionally preserve the original prefix-suffix semantics
//     for patterns like "password " + value.
func (c *SensitiveChecker) hasSensitiveMessage(record logcall.Record) bool {
	if !record.HasMessage() {
		return false
	}

	for _, fragment := range staticMessageFragments(record.Message) {
		if c.hasSensitiveAssignmentPattern(fragment) {
			return true
		}
	}

	if !record.Message.HasDynamicParts() {
		return false
	}

	prefix := record.Message.StaticPrefix()
	if prefix == "" {
		return false
	}

	words := splitWords(prefix)
	if len(words) == 0 {
		return false
	}

	return c.hasSensitiveSuffix(words)
}

func (c *SensitiveChecker) hasSensitiveAssignmentPattern(s string) bool {
	if s == "" {
		return false
	}

	lower := strings.ToLower(s)
	words, ends := splitWordsWithEnds(lower)
	if len(words) == 0 {
		return false
	}

	for _, sequence := range c.keywords {
		if len(sequence) == 0 || len(sequence) > len(words) {
			continue
		}

		for i := 0; i <= len(words)-len(sequence); i++ {
			if !matchesAt(words[i:], sequence) {
				continue
			}

			end := ends[i+len(sequence)-1]
			tail := strings.TrimLeftFunc(lower[end:], func(r rune) bool {
				return unicode.IsSpace(r) || r == '"' || r == '\'' || r == '`'
			})

			if strings.HasPrefix(tail, ":") || strings.HasPrefix(tail, "=") {
				return true
			}
		}
	}

	return false
}

func (c *SensitiveChecker) containsSensitiveToken(s string) bool {
	words := splitWords(s)
	if len(words) == 0 {
		return false
	}

	return c.containsSensitiveSequence(words)
}

func (c *SensitiveChecker) containsSensitiveSequence(words []string) bool {
	for _, sequence := range c.keywords {
		if containsWordSequence(words, sequence) {
			return true
		}
	}

	return false
}

func (c *SensitiveChecker) hasSensitiveSuffix(words []string) bool {
	for _, sequence := range c.keywords {
		if hasWordSequenceSuffix(words, sequence) {
			return true
		}
	}

	return false
}

// splitWords splits a string into lowercase words, removing non-alphanumeric delimiters.
// This is used for splitting user input (message text, field keys) into comparable words.
func splitWords(s string) []string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return nil
	}

	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func splitWordsWithEnds(s string) ([]string, []int) {
	if s == "" {
		return nil, nil
	}

	words := make([]string, 0, 8)
	ends := make([]int, 0, 8)

	start := -1
	for i, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start == -1 {
				start = i
			}
			continue
		}

		if start != -1 {
			words = append(words, s[start:i])
			ends = append(ends, i)
			start = -1
		}
	}

	if start != -1 {
		words = append(words, s[start:])
		ends = append(ends, len(s))
	}

	return words, ends
}

// containsWordSequence checks if the sequence appears anywhere in words.
func containsWordSequence(words, sequence []string) bool {
	if len(words) == 0 || len(sequence) == 0 || len(sequence) > len(words) {
		return false
	}

	for i := 0; i <= len(words)-len(sequence); i++ {
		if matchesAt(words[i:], sequence) {
			return true
		}
	}

	return false
}

// hasWordSequenceSuffix checks if the sequence appears at the end of words.
func hasWordSequenceSuffix(words, sequence []string) bool {
	if len(words) == 0 || len(sequence) == 0 || len(sequence) > len(words) {
		return false
	}

	return matchesAt(words[len(words)-len(sequence):], sequence)
}

// matchesAt checks if words starts with sequence.
func matchesAt(words, sequence []string) bool {
	if len(sequence) > len(words) {
		return false
	}

	for i := range sequence {
		if words[i] != sequence[i] {
			return false
		}
	}

	return true
}
