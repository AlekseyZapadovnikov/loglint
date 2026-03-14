package rules

import (
	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
)

// defaultSensitiveChecker is a singleton checker with default configuration.
// It is created once at package initialization and reused for all CheckSensitive calls.
var defaultSensitiveChecker = NewSensitiveChecker(config.SensitiveConfig{})

// CheckSensitive checks for sensitive data using default configuration.
// This function is kept for backward compatibility.
// For custom configuration, use NewSensitiveChecker directly.
func CheckSensitive(record logcall.Record) []Violation {
	return defaultSensitiveChecker.Check(record)
}
