package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type GameConfig struct {
	Content map[string]map[string]interface{} `yaml:"content"`
}

type Thing struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Cost        map[string]float64 `yaml:"cost"`
	Effects     []Effect           `yaml:"effects"`
	Initial     float64            `yaml:"initial"`
	Reqs        []string           `yaml:"reqs"`
}

type Prestige struct {
	Thing
}

type Resource struct {
	Thing
}

type Upgrade struct {
	Thing
}

type Building struct {
	Thing
	CostIncrease float64 `yaml:"cost_increase"`
}
type Achievement struct {
	Thing
	Levels []AchievementLevel `yaml:"levels"`
}

type AchievementLevel struct {
	Level     int                `yaml:"level"`
	Condition string             `yaml:"condition"`
	Rewards   map[string]float64 `yaml:"rewards"`
}

type Shiny struct {
	Thing
	Frequency float64 `yaml:"frequency"`
	Duration  float64 `yaml:"duration"`
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
