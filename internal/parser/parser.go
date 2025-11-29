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

func ParseConfig(configPath string) (*Config) {
	data, readErr := os.ReadFile(configPath)

	if readErr != nil {
		panic(readErr)
	}

	var config Config

	marshErr := yaml.Unmarshal(data, &config)

	if marshErr != nil {
		panic(marshErr)
	}

	return &config
}
