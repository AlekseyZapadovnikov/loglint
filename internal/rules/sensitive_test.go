package rules_test

import (
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/logcall"
	"github.com/AlekseyZapadovnikov/loglint/internal/rules"
)

func TestSensitiveChecker_DefaultBehavior(t *testing.T) {
	t.Parallel()

	checker := rules.NewSensitiveChecker(config.SensitiveConfig{})

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
		wantMessage   string
	}{
		{
			name:          "dynamic password message is rejected",
			record:        recordWithDynamicMessage("password: "),
			wantViolation: true,
			wantMessage:   sensitiveMessageViolation,
		},
		{
			name:          "static message with token word is allowed",
			record:        recordWithStaticMessage("token validated"),
			wantViolation: false,
		},
		{
			name:          "structured sensitive field is rejected",
			record:        recordWithField("token"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "substring in field key does not trigger",
			record:        recordWithField("stoken_value"),
			wantViolation: false,
		},
		{
			name:          "empty record is ignored",
			record:        logcall.Record{},
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checker.Check(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSensitive, tt.wantMessage)
		})
	}
}

func TestSensitiveChecker_MessageSemantics(t *testing.T) {
	t.Parallel()

	checker := rules.NewSensitiveChecker(config.SensitiveConfig{})

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
	}{
		{
			name:          "colon with sensitive token at suffix triggers",
			record:        recordWithDynamicMessage("password: "),
			wantViolation: true,
		},
		{
			name:          "equals with sensitive token at suffix triggers",
			record:        recordWithDynamicMessage("token = "),
			wantViolation: true,
		},
		{
			name:          "sensitive token without delimiter still triggers",
			record:        recordWithDynamicMessage("password "),
			wantViolation: true,
		},
		{
			name:          "sensitive token not at suffix does not trigger",
			record:        recordWithDynamicMessage("password value: "),
			wantViolation: false,
		},
		{
			name:          "static message is not checked by message matcher",
			record:        recordWithStaticMessage("password: value"),
			wantViolation: false,
		},
		{
			name:          "multi word keyword at suffix triggers",
			record:        recordWithDynamicMessage("api_key: "),
			wantViolation: true,
		},
		{
			name:          "just keyword in dynamic prefix triggers",
			record:        recordWithDynamicMessage("password"),
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checker.Check(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSensitive, sensitiveMessageViolation)
		})
	}
}

func TestSensitiveChecker_ExtraKeywords(t *testing.T) {
	t.Parallel()

	checker := rules.NewSensitiveChecker(config.SensitiveConfig{
		ExtraKeywords: []string{"client_secret", "private_key"},
	})

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
		wantMessage   string
	}{
		{
			name:          "custom keyword in field is rejected",
			record:        recordWithField("client_secret"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "custom keyword in message is rejected",
			record:        recordWithDynamicMessage("private_key: "),
			wantViolation: true,
			wantMessage:   sensitiveMessageViolation,
		},
		{
			name:          "default keyword still works",
			record:        recordWithField("password"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "non sensitive field is allowed",
			record:        recordWithField("user_id"),
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checker.Check(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSensitive, tt.wantMessage)
		})
	}
}

func TestSensitiveChecker_ReplaceDefaults(t *testing.T) {
	t.Parallel()

	checker := rules.NewSensitiveChecker(config.SensitiveConfig{
		ExtraKeywords:   []string{"custom_secret"},
		ReplaceDefaults: true,
	})

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
	}{
		{
			name:          "custom keyword works",
			record:        recordWithField("custom_secret"),
			wantViolation: true,
		},
		{
			name:          "default password keyword is not checked",
			record:        recordWithField("password"),
			wantViolation: false,
		},
		{
			name:          "default token keyword is not checked",
			record:        recordWithField("token"),
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checker.Check(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSensitive, sensitiveFieldViolationMessage)
		})
	}
}

func TestSensitiveChecker_KeywordNormalization(t *testing.T) {
	t.Parallel()

	checker := rules.NewSensitiveChecker(config.SensitiveConfig{
		ExtraKeywords: []string{"client_secret"},
	})

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
		wantMessage   string
	}{
		{
			name:          "snake case in field matches",
			record:        recordWithField("client_secret"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "kebab case in field matches",
			record:        recordWithField("client-secret"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "space separated field matches",
			record:        recordWithField("client secret"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "uppercase field matches",
			record:        recordWithField("CLIENT_SECRET"),
			wantViolation: true,
			wantMessage:   sensitiveFieldViolationMessage,
		},
		{
			name:          "space separated message matches",
			record:        recordWithDynamicMessage("client secret: "),
			wantViolation: true,
			wantMessage:   sensitiveMessageViolation,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checker.Check(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSensitive, tt.wantMessage)
		})
	}
}

func TestSensitiveChecker_EmptyAndDuplicateKeywords(t *testing.T) {
	t.Parallel()

	checker := rules.NewSensitiveChecker(config.SensitiveConfig{
		ExtraKeywords: []string{"", "valid_key", "", "valid_key", " "},
	})

	tests := []struct {
		name          string
		record        logcall.Record
		wantViolation bool
	}{
		{
			name:          "normalized valid keyword works",
			record:        recordWithField("valid_key"),
			wantViolation: true,
		},
		{
			name:          "non listed field is allowed",
			record:        recordWithField("random_field"),
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checker.Check(tt.record)

			if !tt.wantViolation {
				requireViolationCount(t, got, 0)
				return
			}

			requireSingleViolation(t, got, rules.RuleSensitive, sensitiveFieldViolationMessage)
		})
	}
}
