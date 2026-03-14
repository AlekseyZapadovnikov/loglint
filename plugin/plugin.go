package plugin

import (
	"fmt"

	"github.com/AlekseyZapadovnikov/loglint/internal/analyzer"
	"github.com/AlekseyZapadovnikov/loglint/internal/config"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin(analyzer.Name, New)
}

// New creates a golangci-lint module plugin instance from linter settings.
// Settings are decoded with config.DecodePluginSettings.
func New(settings any) (register.LinterPlugin, error) {
	cfg, err := parseSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	return &Plugin{cfg: cfg}, nil
}

// Plugin is the golangci-lint module plugin implementation.
type Plugin struct {
	cfg config.Config
}

var _ register.LinterPlugin = (*Plugin)(nil)

// BuildAnalyzers builds analyzers for the configured plugin instance.
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	a, err := analyzer.New(p.cfg)
	if err != nil {
		return nil, err
	}
	return []*analysis.Analyzer{a}, nil
}

// GetLoadMode returns the required load mode for this plugin.
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// parseSettings parses the settings from golangci-lint configuration.
func parseSettings(settings any) (config.Config, error) {
	return config.DecodePluginSettings(settings)
}
