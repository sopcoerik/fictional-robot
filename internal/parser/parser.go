package parser

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Command string `yaml:"command"`
	Port int `yaml:"port"`
	DependsOn []string `yaml:"depends_on,omitempty"`
}

type Config struct {
	Services map[string]Service 
}

func ParseConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
