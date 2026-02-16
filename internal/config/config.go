package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type FileConfig struct {
	Start                    string   `yaml:"start"`
	End                      string   `yaml:"end"`
	Base                     string   `yaml:"base"`
	Head                     string   `yaml:"head"`
	AllowNesting             bool     `yaml:"allow_nesting"`
	AllowBoundaryWithOutside bool     `yaml:"allow_boundary_with_outside"`
	JSON                     bool     `yaml:"json"`
	Include                  []string `yaml:"include"`
	Exclude                  []string `yaml:"exclude"`
}

func Load(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}
