package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var defaultConfigNames = []string{
	".loglint.yml",
	".loglint.yaml",
	"loglint.yml",
	"loglint.yaml",
}

// LoadStandaloneConfig searches for a standalone config file in dir and loads it.
// If no config file is found, it returns DefaultConfig and an empty path.
func LoadStandaloneConfig(dir string) (Config, string, error) {
	for _, name := range defaultConfigNames {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return DefaultConfig(), "", fmt.Errorf("stat config file %q: %w", path, err)
		}

		if info.IsDir() {
			continue
		}

		cfg, err := LoadFile(path)
		if err != nil {
			return DefaultConfig(), "", err
		}

		return cfg, path, nil
	}

	return DefaultConfig(), "", nil
}

// LoadFile loads configuration from a YAML file using the shared config schema.
func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("read config file %q: %w", path, err)
	}

	return decodeYAML(path, data)
}

func decodeYAML(path string, data []byte) (Config, error) {
	cfg := DefaultConfig()

	if len(bytes.TrimSpace(data)) == 0 {
		return cfg, nil
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config file %q: %w", path, err)
	}

	cfg = cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("validate config file %q: %w", path, err)
	}

	return cfg, nil
}
