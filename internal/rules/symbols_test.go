package rules_test

import (
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
)

func TestCheckSymbols(t *testing.T) {
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
			name:          "plain message is allowed",
			record:        recordWithStaticMessage("request started"),
			wantViolation: false,
		},
		{
			name:          "equals sign is allowed",
			record:        recordWithStaticMessage("request_id=123"),
			wantViolation: false,
		},
		{
			name:          "exclamation mark is rejected",
			record:        recordWithStaticMessage("request started!"),
			wantViolation: true,
		},
		{
			name:          "question mark is rejected",
			record:        recordWithStaticMessage("request started?"),
			wantViolation: true,
		},
		{
			name:          "ellipsis pattern is rejected",
			record:        recordWithStaticMessage("request started..."),
			wantViolation: true,
		},
		{
			name:          "emoji is rejected",
			record:        recordWithStaticMessage("request started 🚀"),
			wantViolation: true,
		},
		{
			name:          "forbidden symbol in static prefix of dynamic message is rejected",
			record:        recordWithDynamicMessage("status? "),
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := rules.CheckSymbols(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSymbols, symbolsViolationMessage)
		})
	}
}
