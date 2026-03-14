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

// New creates a new linter plugin with the given settings.
// Settings should be a map[string]any with the following structure:
//
//	settings:
//	  enabled_rules:
//	    - all
//	  disabled_rules:
//	    - symbols
//	  sensitive:
//	    extra_keywords:
//	      - client_secret
//	      - private_key
//	    replace_defaults: false
func New(settings any) (register.LinterPlugin, error) {
	cfg, err := parseSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	return &Plugin{cfg: cfg}, nil
}

type Plugin struct {
	cfg config.Config
}

var _ register.LinterPlugin = (*Plugin)(nil)

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	a, err := analyzer.New(p.cfg)
	if err != nil {
		return nil, err
	}
	return []*analysis.Analyzer{a}, nil
}

func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// parseSettings parses the settings from golangci-lint configuration.
func parseSettings(settings any) (config.Config, error) {
	return config.DecodePluginSettings(settings)
}
