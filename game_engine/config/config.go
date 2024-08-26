package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type GameConfig struct {
	Content           map[string]map[string]interface{} `yaml:"content"`
	EffectDefinitions map[string]EffectDefinition       `yaml:"effectDefinitions"`
}

type EffectDefinition struct {
	Type       string `yaml:"type"`
	Expression string `yaml:"expression"`
}

type Effect struct {
	Type       string  `yaml:"type"`
	Target     string  `yaml:"target"`
	Value      float64 `yaml:"value"`
	Expression string  `yaml:"expression"`
	Condition  string  `yaml:"condition"`
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
