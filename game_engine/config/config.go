package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type GameConfig struct {
	Content map[string]map[string]interface{} `yaml:"content"`
}

type CustomContent struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Properties  map[string]interface{} `yaml:"properties"`
}

type Prestige struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Cost        map[string]float64 `yaml:"cost"`
	Effects     []EffectConfig     `yaml:"effects"`
}

type Resource struct {
	Initial float64 `yaml:"initial"`
}

type Upgrade struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Cost        map[string]float64 `yaml:"cost"`
	Effects     []EffectConfig     `yaml:"effects"`
}

type Building struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Cost        map[string]float64 `yaml:"cost"`
	Effects     []EffectConfig     `yaml:"effects"`
	Tier        int                `yaml:"tier"`
	Initial     int                `yaml:"initial"`
}
type Achievement struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Condition   string             `yaml:"condition"`
	Levels      []AchievementLevel `yaml:"levels"`
}

type AchievementLevel struct {
	Level     int                `yaml:"level"`
	Condition string             `yaml:"condition"`
	Rewards   map[string]float64 `yaml:"rewards"`
}

type Shiny struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Frequency   float64        `yaml:"frequency"`
	Duration    float64        `yaml:"duration"`
	Effects     []EffectConfig `yaml:"effects"`
}

type EffectConfig struct {
	Type       string  `yaml:"type"`
	Target     string  `yaml:"target"`
	Value      float64 `yaml:"value"`
	Expression string  `yaml:"expression"`
}

type Template struct {
	Type        string                 `yaml:"type"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Tier        int                    `yaml:"tier"`
	Cost        map[string]float64     `yaml:"cost"`
	Effects     []EffectConfig         `yaml:"effects"`
	Content     map[string]interface{} `yaml:"content"`
}

type Include struct {
	Name         string                 `yaml:"name"`
	Tier         int                    `yaml:"tier"`
	Templates    []string               `yaml:"templates"`
	Replacements map[string]interface{} `yaml:"replacements"`
	Parameters   []string               `yaml:"parameters"`
	Content      string                 `yaml:"content"`
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

	// err = config.ProcessTemplatesAndIncludes()
	// if err != nil {
	// 	return nil, err
	// }

	return &config, nil
}

// func (gc *GameConfig) ProcessTemplatesAndIncludes() error {
// 	for name, include := range gc.Includes {
// 		for _, templateName := range include.Templates {
// 			template, ok := gc.Templates[templateName]
// 			if !ok {
// 				return fmt.Errorf("template not found: %s", templateName)
// 			}

// 			switch template.Type {
// 			case "upgrade":
// 				upgrade := Upgrade{
// 					Name:        name,
// 					Description: template.Description,
// 					Cost:        template.Cost,
// 					Effects:     template.Effects,
// 				}
// 				for k, v := range include.Replacements {
// 					switch k {
// 					case "description":
// 						upgrade.Description = v.(string)
// 					case "cost":
// 						upgrade.Cost = v.(map[string]float64)
// 					}
// 				}
// 				gc.Upgrades[name] = upgrade
// 			case "building":
// 				building := Building{
// 					Name:        name,
// 					Description: template.Description,
// 					Cost:        template.Cost,
// 					Effects:     template.Effects,
// 					Tier:        template.Tier,
// 				}
// 				for k, v := range include.Replacements {
// 					switch k {
// 					case "description":
// 						building.Description = v.(string)
// 					case "cost":
// 						building.Cost = v.(map[string]float64)
// 					case "tier":
// 						building.Tier = v.(int)
// 					}
// 				}
// 				gc.Buildings[name] = building
// 			}
// 		}
// 	}
// 	return nil
// }
