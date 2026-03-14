package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlekseyZapadovnikov/loglint/internal/ruleid"
)

// Config is the root configuration for the analyzer.
type Config struct {
	EnabledRules  *[]string       `json:"enabled_rules" yaml:"enabled_rules"`
	DisabledRules []string        `json:"disabled_rules" yaml:"disabled_rules"`
	Sensitive     SensitiveConfig `json:"sensitive" yaml:"sensitive"`
}

// SensitiveConfig holds configuration for the sensitive rule.
type SensitiveConfig struct {
	// ExtraKeywords adds additional sensitive keywords to the default set.
	// Keywords are normalized (lowercase, split by delimiters).
	// Empty strings and duplicates are ignored.
	ExtraKeywords []string `json:"extra_keywords" yaml:"extra_keywords"`

	// ReplaceDefaults when true replaces the default keyword set entirely.
	// When false (default), ExtraKeywords are added to the defaults.
	ReplaceDefaults bool `json:"replace_defaults" yaml:"replace_defaults"`
}

// DefaultSensitiveKeywords stores built-in sensitive keywords in canonical form.
var DefaultSensitiveKeywords = []string{
	"password",
	"passwd",
	"pwd",
	"secret",
	"token",
	"api_key",
	"apikey",
	"access_token",
	"refresh_token",
	"authorization",
	"bearer",
	"cookie",
	"session",
	"session_id",
	"credential",
	"credentials",
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		EnabledRules:  nil,
		DisabledRules: nil,
		Sensitive: SensitiveConfig{
			ExtraKeywords:   nil,
			ReplaceDefaults: false,
		},
	}
}

// Normalize returns a normalized copy of the configuration.
func (c Config) Normalize() Config {
	normalized := DefaultConfig()

	if c.EnabledRules != nil {
		enabled := normalizeRuleNames(*c.EnabledRules)
		normalized.EnabledRules = &enabled
	}

	normalized.DisabledRules = normalizeRuleNames(c.DisabledRules)
	normalized.Sensitive = c.Sensitive.Normalize()

	return normalized
}

// Normalize returns a normalized copy of SensitiveConfig:
//   - converts to lowercase
//   - splits by delimiters (underscore, hyphen, space)
//   - removes empty strings and duplicates
//   - sorts for stable output
func (c SensitiveConfig) Normalize() SensitiveConfig {
	normalized := make([]string, 0, len(c.ExtraKeywords))
	seen := make(map[string]struct{})

	for _, kw := range c.ExtraKeywords {
		canonical := NormalizeKeyword(kw)
		if canonical == "" {
			continue
		}

		if _, exists := seen[canonical]; exists {
			continue
		}
		seen[canonical] = struct{}{}
		normalized = append(normalized, canonical)
	}

	sort.Strings(normalized)

	return SensitiveConfig{
		ExtraKeywords:   normalized,
		ReplaceDefaults: c.ReplaceDefaults,
	}
}

// Validate validates SensitiveConfig invariants.
func (c SensitiveConfig) Validate() error {
	return nil
}

// Validate validates the entire configuration.
func (c Config) Validate() error {
	if _, err := c.ResolveRules(); err != nil {
		return err
	}

	if err := c.Sensitive.Validate(); err != nil {
		return fmt.Errorf("sensitive config: %w", err)
	}
	return nil
}

// ResolveRules resolves the active rule set using enabled_rules and disabled_rules.
func (c Config) ResolveRules() ([]ruleid.ID, error) {
	cfg := c.Normalize()

	active := make(map[ruleid.ID]struct{}, len(ruleid.OrderedRules()))

	if cfg.EnabledRules == nil {
		for _, id := range ruleid.OrderedRules() {
			active[id] = struct{}{}
		}
	} else {
		for i, name := range *cfg.EnabledRules {
			if ruleid.NormalizeName(name) == string(ruleid.All) {
				for _, id := range ruleid.OrderedRules() {
					active[id] = struct{}{}
				}
				continue
			}

			id, ok := ruleid.ParseRule(name)
			if !ok {
				return nil, fmt.Errorf(
					"enabled_rules[%d]: unknown rule %q (allowed: %s)",
					i,
					name,
					strings.Join(ruleid.AllowedEnabledNames(), ", "),
				)
			}

			active[id] = struct{}{}
		}
	}

	for i, name := range cfg.DisabledRules {
		if ruleid.NormalizeName(name) == string(ruleid.All) {
			return nil, fmt.Errorf(
				"disabled_rules[%d]: %q is not allowed (allowed: %s)",
				i,
				name,
				strings.Join(ruleid.AllowedRuleNames(), ", "),
			)
		}

		id, ok := ruleid.ParseRule(name)
		if !ok {
			return nil, fmt.Errorf(
				"disabled_rules[%d]: unknown rule %q (allowed: %s)",
				i,
				name,
				strings.Join(ruleid.AllowedRuleNames(), ", "),
			)
		}

		delete(active, id)
	}

	resolved := make([]ruleid.ID, 0, len(active))
	for _, id := range ruleid.OrderedRules() {
		if _, ok := active[id]; ok {
			resolved = append(resolved, id)
		}
	}

	return resolved, nil
}

// NormalizeKeyword converts a keyword to its canonical form.
// It normalizes the keyword:
//   - converts to lowercase
//   - splits by common delimiters (_, -, space)
//   - removes empty parts
//   - joins with underscore
//
// This ensures "client_secret", "client-secret", "client secret" are treated identically.
func NormalizeKeyword(keyword string) string {
	words := SplitKeywordToWords(keyword)
	if len(words) == 0 {
		return ""
	}
	return strings.Join(words, "_")
}

// SplitKeywordToWords splits a keyword into normalized words.
// This is the core normalization function used by both config and rules.
// It handles: lowercase conversion, splitting by _/-/space, filtering empty parts.
func SplitKeywordToWords(keyword string) []string {
	s := strings.ToLower(strings.TrimSpace(keyword))
	if s == "" {
		return nil
	}

	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}

	return parts
}

func normalizeRuleNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(names))
	for _, name := range names {
		normalized = append(normalized, ruleid.NormalizeName(name))
	}

	return normalized
}
