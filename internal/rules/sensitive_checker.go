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
		// Use only extra keywords
		keywords = buildKeywordsFromCanonical(cfg.ExtraKeywords)
	} else {
		// Merge defaults with extra keywords
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
		// Filter empty strings that might result from double underscores
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
// Message semantics (preserved from original implementation):
//  1. Only checks messages with dynamic parts (concatenation, formatting)
//  2. Only checks the static prefix of the message
//  3. Looks for sensitive keyword at the END of the prefix (suffix matching)
func (c *SensitiveChecker) hasSensitiveMessage(record logcall.Record) bool {
	// Only check messages with dynamic parts
	if !record.HasMessage() || !record.Message.HasDynamicParts() {
		return false
	}

	prefix := record.Message.StaticPrefix()
	if prefix == "" {
		return false
	}

	// Split prefix into words
	words := splitWords(prefix)
	if len(words) == 0 {
		return false
	}

	// Check if sensitive keyword appears at the end of the prefix
	return c.hasSensitiveSuffix(words)
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
