package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type GameConfig struct {
	Content map[string]map[string]interface{} `yaml:"content"`
}

func LoadConfig(filename string) (*GameConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config GameConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
