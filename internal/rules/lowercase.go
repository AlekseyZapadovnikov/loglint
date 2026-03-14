package rules

import (
	"unicode"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
)

// CheckLowercase reports messages that start with an uppercase letter.
func CheckLowercase(record logcall.Record) []Violation {
	if !record.HasMessage() {
		return nil
	}

	text := record.Message.StaticPrefix()
	if text == "" {
		return nil
	}

	first, ok := firstMeaningfulLetter(text)
	if !ok {
		return nil
	}

	if !unicode.IsUpper(first) && !unicode.IsTitle(first) {
		return nil
	}

	return []Violation{
		newViolation(
			RuleLowercase,
			record.Message.Expr,
			"log message must start with a lowercase letter",
		),
	}
}

func firstMeaningfulLetter(s string) (rune, bool) {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return r, true
		}
	}

	return 0, false
}
