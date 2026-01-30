package defaults

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed aliases.yaml
var aliasesYAML []byte

// EmbeddedDefaults contains default configuration embedded in the binary
type EmbeddedDefaults struct {
	Aliases map[string]string `yaml:"aliases"`
}

// Load returns the embedded default configuration
func Load() (*EmbeddedDefaults, error) {
	var defaults EmbeddedDefaults
	if err := yaml.Unmarshal(aliasesYAML, &defaults); err != nil {
		return nil, err
	}
	if defaults.Aliases == nil {
		defaults.Aliases = make(map[string]string)
	}
	return &defaults, nil
}

// GetAlias returns an embedded default alias, or empty string if not found
func GetAlias(alias string) string {
	defaults, err := Load()
	if err != nil {
		return ""
	}
	return defaults.Aliases[alias]
}

// ListAliases returns all embedded default aliases
func ListAliases() map[string]string {
	defaults, err := Load()
	if err != nil {
		return make(map[string]string)
	}
	return defaults.Aliases
}
