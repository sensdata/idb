package manager

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	Plugins []PluginEntry `yaml:"plugins"`
}

type PluginEntry struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Url     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

func LoadRegistry(data []byte) (*Registry, error) {
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry yaml: %w", err)
	}

	return &reg, nil
}
