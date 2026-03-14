package rules_test

import (
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
)

func TestCheckEnglish(t *testing.T) {
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
			name:          "dynamic message without static text is ignored",
			record:        recordWithDynamicMessage(""),
			wantViolation: false,
		},
		{
			name:          "latin message is allowed",
			record:        recordWithStaticMessage("request started"),
			wantViolation: false,
		},
		{
			name:          "numbers and punctuation are allowed",
			record:        recordWithStaticMessage("request_id=42"),
			wantViolation: false,
		},
		{
			name:          "cyrillic message is rejected",
			record:        recordWithStaticMessage("запуск сервера"),
			wantViolation: true,
		},
		{
			name:          "non latin static prefix in dynamic message is rejected",
			record:        recordWithDynamicMessage("ошибка: "),
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := rules.CheckEnglish(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleEnglish, englishViolationMessage)
		})
	}
}
