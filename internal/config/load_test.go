package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/AlekseyZapadovnikov/loglint/internal/ruleid"
)

func TestLoadStandaloneConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".loglint.yml")

	content := []byte(`
enabled_rules:
  - all
disabled_rules:
  - symbols
sensitive:
  extra_keywords:
    - client_secret
  replace_defaults: false
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, foundPath, err := config.LoadStandaloneConfig(dir)
	if err != nil {
		t.Fatalf("LoadStandaloneConfig error: %v", err)
	}

	if foundPath != path {
		t.Fatalf("found path: got %q, want %q", foundPath, path)
	}

	rules, err := cfg.ResolveRules()
	if err != nil {
		t.Fatalf("ResolveRules error: %v", err)
	}

	wantRules := []ruleid.ID{
		ruleid.Lowercase,
		ruleid.English,
		ruleid.Sensitive,
	}
	if len(rules) != len(wantRules) {
		t.Fatalf("rules length: got %d, want %d", len(rules), len(wantRules))
	}
	for i := range rules {
		if rules[i] != wantRules[i] {
			t.Fatalf("rules[%d]: got %q, want %q", i, rules[i], wantRules[i])
		}
	}

	if len(cfg.Sensitive.ExtraKeywords) != 1 || cfg.Sensitive.ExtraKeywords[0] != "client_secret" {
		t.Fatalf("unexpected sensitive keywords: %#v", cfg.Sensitive.ExtraKeywords)
	}
}
