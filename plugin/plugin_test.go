package plugin

import (
	"strings"
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/ruleid"
)

func TestParseSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		settings    any
		wantRules   []ruleid.ID
		wantConfig  config.Config
		errContains string
		expectError bool
	}{
		{
			name:      "nil settings returns default config",
			settings:  nil,
			wantRules: ruleid.OrderedRules(),
		},
		{
			name:      "empty map returns default config",
			settings:  map[string]any{},
			wantRules: ruleid.OrderedRules(),
		},
		{
			name:        "invalid settings type returns error",
			settings:    "invalid",
			expectError: true,
		},
		{
			name: "plugin settings decode rule selectors",
			settings: map[string]any{
				"enabled_rules":  []any{"all"},
				"disabled_rules": []any{"symbols"},
			},
			wantRules: []ruleid.ID{
				ruleid.Lowercase,
				ruleid.English,
				ruleid.Sensitive,
			},
		},
		{
			name: "sensitive extra_keywords are parsed",
			settings: map[string]any{
				"sensitive": map[string]any{
					"extra_keywords": []any{"client_secret", "private_key"},
				},
			},
			wantRules: ruleid.OrderedRules(),
			wantConfig: config.Config{
				Sensitive: config.SensitiveConfig{
					ExtraKeywords:   []string{"client_secret", "private_key"},
					ReplaceDefaults: false,
				},
			},
		},
		{
			name: "sensitive replace_defaults is parsed",
			settings: map[string]any{
				"sensitive": map[string]any{
					"replace_defaults": true,
				},
			},
			wantRules: ruleid.OrderedRules(),
			wantConfig: config.Config{
				Sensitive: config.SensitiveConfig{
					ExtraKeywords:   []string{},
					ReplaceDefaults: true,
				},
			},
		},
		{
			name: "backward compatible sensitive config is parsed",
			settings: map[string]any{
				"sensitive": map[string]any{
					"extra_keywords":   []any{"custom_key"},
					"replace_defaults": true,
				},
			},
			wantRules: ruleid.OrderedRules(),
			wantConfig: config.Config{
				Sensitive: config.SensitiveConfig{
					ExtraKeywords:   []string{"custom_key"},
					ReplaceDefaults: true,
				},
			},
		},
		{
			name: "invalid sensitive type returns error",
			settings: map[string]any{
				"sensitive": "invalid",
			},
			expectError: true,
		},
		{
			name: "unknown enabled rule returns error",
			settings: map[string]any{
				"enabled_rules": []any{"unknown"},
			},
			expectError: true,
			errContains: `unknown rule "unknown"`,
		},
		{
			name: "all in disabled rules returns error",
			settings: map[string]any{
				"disabled_rules": []any{"all"},
			},
			expectError: true,
			errContains: `"all" is not allowed`,
		},
		{
			name: "invalid extra_keywords type returns error",
			settings: map[string]any{
				"sensitive": map[string]any{
					"extra_keywords": "not a slice",
				},
			},
			expectError: true,
		},
		{
			name: "invalid extra_keywords element type returns error",
			settings: map[string]any{
				"sensitive": map[string]any{
					"extra_keywords": []any{123},
				},
			},
			expectError: true,
		},
		{
			name: "invalid replace_defaults type returns error",
			settings: map[string]any{
				"sensitive": map[string]any{
					"replace_defaults": "not a bool",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseSettings(tt.settings)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotRules, err := got.ResolveRules()
			if err != nil {
				t.Fatalf("ResolveRules error: %v", err)
			}
			if len(gotRules) != len(tt.wantRules) {
				t.Fatalf("rules length: got %d, want %d", len(gotRules), len(tt.wantRules))
			}
			for i := range gotRules {
				if gotRules[i] != tt.wantRules[i] {
					t.Fatalf("rules[%d]: got %q, want %q", i, gotRules[i], tt.wantRules[i])
				}
			}

			if got.Sensitive.ReplaceDefaults != tt.wantConfig.Sensitive.ReplaceDefaults {
				t.Errorf("ReplaceDefaults: got %v, want %v", got.Sensitive.ReplaceDefaults, tt.wantConfig.Sensitive.ReplaceDefaults)
			}

			if len(got.Sensitive.ExtraKeywords) != len(tt.wantConfig.Sensitive.ExtraKeywords) {
				t.Errorf("ExtraKeywords length: got %d, want %d", len(got.Sensitive.ExtraKeywords), len(tt.wantConfig.Sensitive.ExtraKeywords))
				return
			}

			for i := range got.Sensitive.ExtraKeywords {
				if got.Sensitive.ExtraKeywords[i] != tt.wantConfig.Sensitive.ExtraKeywords[i] {
					t.Errorf("ExtraKeywords[%d]: got %q, want %q", i, got.Sensitive.ExtraKeywords[i], tt.wantConfig.Sensitive.ExtraKeywords[i])
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		settings    any
		expectError bool
	}{
		{
			name:     "nil settings creates plugin",
			settings: nil,
		},
		{
			name: "valid settings creates plugin",
			settings: map[string]any{
				"sensitive": map[string]any{
					"extra_keywords": []any{"custom_key"},
				},
			},
		},
		{
			name:        "invalid settings type returns error",
			settings:    "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plugin, err := New(tt.settings)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if plugin == nil {
				t.Fatal("plugin is nil")
			}

			// Verify BuildAnalyzers works
			analyzers, err := plugin.BuildAnalyzers()
			if err != nil {
				t.Fatalf("BuildAnalyzers error: %v", err)
			}

			if len(analyzers) != 1 {
				t.Errorf("expected 1 analyzer, got %d", len(analyzers))
			}

			if analyzers[0].Name != "loglint" {
				t.Errorf("expected analyzer name 'loglint', got %q", analyzers[0].Name)
			}
		})
	}
}
