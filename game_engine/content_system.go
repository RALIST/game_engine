package game_engine

import (
	"fmt"
	"github.com/ralist/game_engine/game_engine/config"
	"log"

	"gopkg.in/yaml.v2"
)

type Thing interface {
}

type ContentSystem struct {
	config       *config.GameConfig
	Resources    map[string]config.Resource
	Buildings    map[string]config.Building
	Upgrades     map[string]config.Upgrade
	Achievements []config.Achievement
	Things       map[string]config.Thing
	Shinies      map[string]config.Shiny
	Prestige     config.Prestige
}

func NewContentSystem(cfg *config.GameConfig) (*ContentSystem, error) {
	cs := &ContentSystem{
		config:       cfg,
		Resources:    make(map[string]config.Resource),
		Buildings:    make(map[string]config.Building),
		Upgrades:     make(map[string]config.Upgrade),
		Achievements: []config.Achievement{},
		Shinies:      make(map[string]config.Shiny),
		Things:       make(map[string]config.Thing),
	}

	err := cs.parseContent()
	log.Println(cs.Things)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func (cs *ContentSystem) parseContent() error {
	for category, content := range cs.config.Content {
		switch category {
		case "resources":
			if err := cs.parseResources(content); err != nil {
				return err
			}
		case "buildings":
			if err := cs.parseBuildings(content); err != nil {
				return err
			}
		case "upgrades":
			if err := cs.parseUpgrades(content); err != nil {
				return err
			}
		case "achievements":
			if err := cs.parseAchievements(content); err != nil {
				return err
			}
		case "shinies":
			if err := cs.parseShinies(content); err != nil {
				return err
			}
		default:
			// Treat any other category as custom content
			if err := cs.parseThings(content); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *ContentSystem) parseThings(content map[string]interface{}) error {
	for name, data := range content {
		var thing config.Thing
		if err := remarshal(data, &thing); err != nil {
			return fmt.Errorf("error parsing resource %s: %v", name, err)
		}
		cs.Things[name] = thing
	}
	return nil
}

func (cs *ContentSystem) parseResources(content map[string]interface{}) error {
	for name, data := range content {
		var resource config.Resource
		if err := remarshal(data, &resource); err != nil {
			return fmt.Errorf("error parsing resource %s: %v", name, err)
		}
		cs.Resources[name] = resource
	}
	return nil
}

func (cs *ContentSystem) parseBuildings(content map[string]interface{}) error {
	for name, data := range content {
		var building config.Building
		if err := remarshal(data, &building); err != nil {
			return fmt.Errorf("error parsing building %s: %v", name, err)
		}
		cs.Buildings[name] = building
	}
	return nil
}

func (cs *ContentSystem) parseUpgrades(content map[string]interface{}) error {
	for name, data := range content {
		var upgrade config.Upgrade
		if err := remarshal(data, &upgrade); err != nil {
			return fmt.Errorf("error parsing upgrade %s: %v", name, err)
		}
		cs.Upgrades[name] = upgrade
	}
	return nil
}

func (cs *ContentSystem) parseAchievements(content map[string]interface{}) error {
	for _, data := range content {
		var achievement config.Achievement
		if err := remarshal(data, &achievement); err != nil {
			return fmt.Errorf("error parsing achievement: %v", err)
		}
		cs.Achievements = append(cs.Achievements, achievement)
	}
	return nil
}

func (cs *ContentSystem) parseShinies(content map[string]interface{}) error {
	for name, data := range content {
		var shiny config.Shiny
		if err := remarshal(data, &shiny); err != nil {
			return fmt.Errorf("error parsing shiny %s: %v", name, err)
		}
		cs.Shinies[name] = shiny
	}
	return nil
}

func remarshal(in interface{}, out interface{}) error {
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

type ContentType string

func (cs *ContentSystem) GetContent(category, name string) (map[string]interface{}, error) {
	if categoryContent, ok := cs.config.Content[category]; ok {
		if content, ok := categoryContent[name]; ok {
			return content.(map[string]interface{}), nil
		}
	}
	return nil, fmt.Errorf("content not found: %s in category %s", name, category)
}

func (cs *ContentSystem) GetAllContent(category string) map[string]interface{} {
	return cs.config.Content[category]
}

func (cs *ContentSystem) GetCategories() []string {
	categories := make([]string, 0, len(cs.config.Content))
	for category := range cs.config.Content {
		categories = append(categories, category)
	}
	return categories
}

func (cs *ContentSystem) GetResource(name string) (config.Resource, error) {
	if resource, ok := cs.Resources[name]; ok {
		return resource, nil
	}
	return config.Resource{}, fmt.Errorf("resource not found: %s", name)
}

func (cs *ContentSystem) GetBuilding(name string) (config.Building, error) {
	if building, ok := cs.Buildings[name]; ok {
		return building, nil
	}
	return config.Building{}, fmt.Errorf("building not found: %s", name)
}

func (cs *ContentSystem) GetUpgrade(name string) (config.Upgrade, error) {
	if upgrade, ok := cs.Upgrades[name]; ok {
		return upgrade, nil
	}
	return config.Upgrade{}, fmt.Errorf("upgrade not found: %s", name)
}

func (cs *ContentSystem) GetAchievement(name string) (config.Achievement, error) {
	for _, achievement := range cs.Achievements {
		if achievement.Name == name {
			return achievement, nil
		}
	}
	return config.Achievement{}, fmt.Errorf("achievement not found: %s", name)
}

func (cs *ContentSystem) GetShiny(name string) (config.Shiny, error) {
	if shiny, ok := cs.Shinies[name]; ok {
		return shiny, nil
	}
	return config.Shiny{}, fmt.Errorf("shiny not found: %s", name)
}

func (cs *ContentSystem) GetAllResources() map[string]config.Resource {
	return cs.Resources
}

func (cs *ContentSystem) GetAllBuildings() map[string]config.Building {
	return cs.Buildings
}

func (cs *ContentSystem) GetAllUpgrades() map[string]config.Upgrade {
	return cs.Upgrades
}

func (cs *ContentSystem) GetAllAchievements() []config.Achievement {
	return cs.Achievements
}

func (cs *ContentSystem) GetAllShinies() map[string]config.Shiny {
	return cs.Shinies
}
