package rules_test

import (
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
)

func TestCheckLowercase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
	}{
		{
			name:          "missing message is ignored",
			record:        logcall.Record{},
			wantViolation: false,
		},
		{
			name:          "dynamic message without static prefix is ignored",
			record:        recordWithDynamicMessage(""),
			wantViolation: false,
		},
		{
			name:          "lowercase message is allowed",
			record:        recordWithStaticMessage("request started"),
			wantViolation: false,
		},
		{
			name:          "uppercase first letter is rejected",
			record:        recordWithStaticMessage("Request started"),
			wantViolation: true,
		},
		{
			name:          "first meaningful letter after digits is checked",
			record:        recordWithStaticMessage("123 Request started"),
			wantViolation: true,
		},
		{
			name:          "punctuation before lowercase is allowed",
			record:        recordWithStaticMessage("[worker] request started"),
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := rules.CheckLowercase(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleLowercase, lowercaseViolationMessage)
		})
	}
}
