package rules

import (
	"strings"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
)

func CheckSymbols(record logcall.Record) []Violation {
	if !record.HasMessage() {
		return nil
	}

	for _, fragment := range staticMessageFragments(record.Message) {
		if containsForbiddenRune(fragment) || hasForbiddenPunctuationPattern(fragment) {
			return []Violation{
				newViolation(
					RuleSymbols,
					record.Message.Expr,
					"log message must not contain special symbols or emoji",
				),
			}
		}
	}

	return nil
}

func containsForbiddenRune(s string) bool {
	for _, r := range s {
		if isEmojiRune(r) {
			return true
		}

		if isForbiddenPunctuation(r) {
			return true
		}
	}

	return false
}

// We intentionally ban only noisy punctuation and emoji to avoid
// excessive false positives for normal operational log messages.
func isForbiddenPunctuation(r rune) bool {
	switch r {
	case '!', '?', '…':
		return true
	default:
		return false
	}
}

func hasForbiddenPunctuationPattern(s string) bool {
	return strings.Contains(s, "...") ||
		strings.Contains(s, "!!") ||
		strings.Contains(s, "??")
}

func isEmojiRune(r rune) bool {
	switch {
	case r >= 0x1F300 && r <= 0x1FAFF:
		return true
	case r >= 0x2600 && r <= 0x27BF:
		return true
	case r >= 0x2B50 && r <= 0x2B59:
		return true
	default:
		return false
	}
}
