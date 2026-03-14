package config_test

import (
	"strings"
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/ruleid"
)

func TestSensitiveConfigNormalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      config.SensitiveConfig
		expected config.SensitiveConfig
	}{
		{
			name: "empty config",
			cfg: config.SensitiveConfig{
				ExtraKeywords: nil,
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{},
				ReplaceDefaults: false,
			},
		},
		{
			name: "normalizes keywords to lowercase",
			cfg: config.SensitiveConfig{
				ExtraKeywords: []string{"CLIENT_SECRET", "API_KEY"},
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"api_key", "client_secret"},
				ReplaceDefaults: false,
			},
		},
		{
			name: "splits hyphen and space delimiters",
			cfg: config.SensitiveConfig{
				ExtraKeywords: []string{"client-secret", "api key"},
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"api_key", "client_secret"},
				ReplaceDefaults: false,
			},
		},
		{
			name: "filters empty strings",
			cfg: config.SensitiveConfig{
				ExtraKeywords: []string{"", "valid_key", "   ", ""},
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"valid_key"},
				ReplaceDefaults: false,
			},
		},
		{
			name: "removes duplicates",
			cfg: config.SensitiveConfig{
				ExtraKeywords: []string{"client_secret", "client-secret", "CLIENT_SECRET"},
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"client_secret"},
				ReplaceDefaults: false,
			},
		},
		{
			name: "sorts keywords",
			cfg: config.SensitiveConfig{
				ExtraKeywords: []string{"zebra", "alpha", "middle"},
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"alpha", "middle", "zebra"},
				ReplaceDefaults: false,
			},
		},
		{
			name: "preserves replace_defaults flag",
			cfg: config.SensitiveConfig{
				ExtraKeywords:   []string{"key"},
				ReplaceDefaults: true,
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"key"},
				ReplaceDefaults: true,
			},
		},
		{
			name: "handles complex normalization",
			cfg: config.SensitiveConfig{
				ExtraKeywords: []string{
					"  Client-Secret  ",
					"API-KEY",
					"api key",
					"",              // empty
					"CLIENT_SECRET", // duplicate
				},
			},
			expected: config.SensitiveConfig{
				ExtraKeywords:   []string{"api_key", "client_secret"},
				ReplaceDefaults: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.cfg.Normalize()

			if got.ReplaceDefaults != tt.expected.ReplaceDefaults {
				t.Errorf("ReplaceDefaults: got %v, want %v", got.ReplaceDefaults, tt.expected.ReplaceDefaults)
			}

			if len(got.ExtraKeywords) != len(tt.expected.ExtraKeywords) {
				t.Errorf("ExtraKeywords length: got %d, want %d", len(got.ExtraKeywords), len(tt.expected.ExtraKeywords))
				return
			}

			for i := range got.ExtraKeywords {
				if got.ExtraKeywords[i] != tt.expected.ExtraKeywords[i] {
					t.Errorf("ExtraKeywords[%d]: got %q, want %q", i, got.ExtraKeywords[i], tt.expected.ExtraKeywords[i])
				}
			}
		})
	}
}

func TestNormalizeKeyword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keyword  string
		expected string
	}{
		{"client_secret", "client_secret"},
		{"client-secret", "client_secret"},
		{"client secret", "client_secret"},
		{"Client-Secret", "client_secret"},
		{"CLIENT_SECRET", "client_secret"},
		{"  api_key  ", "api_key"},
		{"", ""},
		{"   ", ""},
		{"API KEY", "api_key"},
		{"api--key", "api_key"},
		{"api__key", "api_key"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.keyword, func(t *testing.T) {
			t.Parallel()

			got := config.NormalizeKeyword(tt.keyword)
			if got != tt.expected {
				t.Errorf("config.NormalizeKeyword(%q) = %q, want %q", tt.keyword, got, tt.expected)
			}
		})
	}
}

func TestSplitKeywordToWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		keyword  string
		expected []string
	}{
		{"client_secret", []string{"client", "secret"}},
		{"client-secret", []string{"client", "secret"}},
		{"client secret", []string{"client", "secret"}},
		{"Client-Secret", []string{"client", "secret"}},
		{"CLIENT_SECRET", []string{"client", "secret"}},
		{"  api_key  ", []string{"api", "key"}},
		{"", nil},
		{"   ", nil},
		{"single", []string{"single"}},
		{"api  key", []string{"api", "key"}}, // double space
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.keyword, func(t *testing.T) {
			t.Parallel()

			got := config.SplitKeywordToWords(tt.keyword)
			if len(got) != len(tt.expected) {
				t.Errorf("config.SplitKeywordToWords(%q) = %v, want %v", tt.keyword, got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("config.SplitKeywordToWords(%q)[%d] = %q, want %q", tt.keyword, i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	if cfg.Sensitive.ReplaceDefaults {
		t.Error("ReplaceDefaults should be false by default")
	}

	if len(cfg.Sensitive.ExtraKeywords) != 0 {
		t.Error("ExtraKeywords should be empty by default")
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should be valid, got error: %v", err)
	}
}

func TestConfigResolveRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cfg         config.Config
		want        []ruleid.ID
		errContains string
	}{
		{
			name: "no config enables all rules",
			cfg:  config.DefaultConfig(),
			want: []ruleid.ID{
				ruleid.Lowercase,
				ruleid.English,
				ruleid.Symbols,
				ruleid.Sensitive,
			},
		},
		{
			name: "enabled_rules subset",
			cfg: config.Config{
				EnabledRules: stringSlicePtr("symbols", "lowercase"),
			},
			want: []ruleid.ID{
				ruleid.Lowercase,
				ruleid.Symbols,
			},
		},
		{
			name: "enabled_rules all",
			cfg: config.Config{
				EnabledRules: stringSlicePtr("all"),
			},
			want: []ruleid.ID{
				ruleid.Lowercase,
				ruleid.English,
				ruleid.Symbols,
				ruleid.Sensitive,
			},
		},
		{
			name: "enabled_rules all with disabled_rules",
			cfg: config.Config{
				EnabledRules:  stringSlicePtr("all"),
				DisabledRules: []string{"symbols"},
			},
			want: []ruleid.ID{
				ruleid.Lowercase,
				ruleid.English,
				ruleid.Sensitive,
			},
		},
		{
			name: "unknown enabled rule returns error",
			cfg: config.Config{
				EnabledRules: stringSlicePtr("unknown"),
			},
			errContains: `unknown rule "unknown"`,
		},
		{
			name: "all in disabled_rules returns error",
			cfg: config.Config{
				DisabledRules: []string{"all"},
			},
			errContains: `"all" is not allowed`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.cfg.ResolveRules()
			if tt.errContains != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("resolved rules length: got %d, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("resolved rules[%d]: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func stringSlicePtr(values ...string) *[]string {
	s := append([]string(nil), values...)
	return &s
}
