package config

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DecodePluginSettings decodes golangci-lint plugin settings into the shared config schema.
func DecodePluginSettings(settings any) (Config, error) {
	cfg := DefaultConfig()

	if settings == nil {
		return cfg, nil
	}

	settingsMap, ok := settings.(map[string]any)
	if !ok {
		return cfg, fmt.Errorf("settings must be a map, got %T", settings)
	}

	data, err := json.Marshal(settingsMap)
	if err != nil {
		return cfg, fmt.Errorf("marshal settings: %w", err)
	}

	decoded, err := decodeJSON(data)
	if err != nil {
		return cfg, err
	}

	return decoded, nil
}

func decodeJSON(data []byte) (Config, error) {
	cfg := DefaultConfig()

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}

	cfg = cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}
