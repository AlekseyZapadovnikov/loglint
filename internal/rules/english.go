package rules

import (
	"unicode"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
)

// CheckEnglish reports messages that contain non-Latin letters.
func CheckEnglish(record logcall.Record) []Violation {
	if !record.HasMessage() {
		return nil
	}

	for _, fragment := range staticMessageFragments(record.Message) {
		if containsNonLatinLetter(fragment) {
			return []Violation{
				newViolation(
					RuleEnglish,
					record.Message.Expr,
					"log message must contain only English text",
				),
			}
		}
	}

	return nil
}

func containsNonLatinLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}

		if !unicode.In(r, unicode.Latin) {
			return true
		}
	}

	return false
}
