package game_engine

import (
	"fmt"
	"github.com/ralist/game_engine/game_engine/config"
)

type ContentItem struct {
	Type        string                 `yaml:"type"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Cost        map[string]float64     `yaml:"cost"`
	Effects     []config.Effect        `yaml:"effects"`
	Initial     int                    `yaml:"initial"`
	Reqs        []string               `yaml:"reqs"`
	Properties  map[string]interface{} `yaml:"properties"`
}

type ContentSystem struct {
	content           map[string]map[string]ContentItem
	effectDefinitions map[string]config.EffectDefinition
	pluginSystem      *PluginSystem
}

func NewContentSystem(cfg *config.GameConfig) (*ContentSystem, error) {
	cs := &ContentSystem{
		content:           make(map[string]map[string]ContentItem),
		effectDefinitions: cfg.EffectDefinitions,
		pluginSystem:      NewPluginSystem(),
	}

	err := cs.parseContent(cfg.Content)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (cs *ContentSystem) parseContent(content map[string]map[string]interface{}) error {
	for category, items := range content {
		cs.content[category] = make(map[string]ContentItem)
		for name, data := range items {
			convertedData := convertMapInterfaceToMapString(data.(map[interface{}]interface{}))
			item, err := cs.createContentItem(name, convertedData)
			if err != nil {
				return err
			}
			cs.content[category][name] = item
		}
	}
	return nil
}

func convertMapInterfaceToMapString(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		switch key := k.(type) {
		case string:
			switch value := v.(type) {
			case map[interface{}]interface{}:
				result[key] = convertMapInterfaceToMapString(value)
			case []interface{}:
				result[key] = convertSliceInterfaceToSliceString(value)
			default:
				result[key] = v
			}
		}
	}
	return result
}

func convertSliceInterfaceToSliceString(slice []interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		switch value := v.(type) {
		case map[interface{}]interface{}:
			result[i] = convertMapInterfaceToMapString(value)
		case []interface{}:
			result[i] = convertSliceInterfaceToSliceString(value)
		default:
			result[i] = v
		}
	}
	return result
}

func (cs *ContentSystem) createContentItem(name string, data map[string]interface{}) (ContentItem, error) {
	var item ContentItem
	item.Type = name

	if typeStr, ok := data["name"].(string); ok {
		item.Name = typeStr
	}

	if typeStr, ok := data["type"].(string); ok {
		item.Type = typeStr
	}

	if desc, ok := data["description"].(string); ok {
		item.Description = desc
	}

	if cost, ok := data["cost"].(map[string]interface{}); ok {
		item.Cost = make(map[string]float64)
		for resource, amount := range cost {
			item.Cost[resource] = float64(amount.(int))
		}
	}

	if effects, ok := data["effects"].([]interface{}); ok {
		item.Effects = make([]config.Effect, 0)
		for _, effect := range effects {
			if effectMap, ok := effect.(map[string]interface{}); ok {
				newEffect := config.Effect{}
				if typeStr, ok := effectMap["type"].(string); ok {
					newEffect.Type = typeStr
				}
				if target, ok := effectMap["target"].(string); ok {
					newEffect.Target = target
				}
				if value, ok := effectMap["value"].(float64); ok {
					newEffect.Value = value
				}
				if expression, ok := effectMap["expression"].(string); ok {
					newEffect.Expression = expression
				}
				if condition, ok := effectMap["condition"].(string); ok {
					newEffect.Condition = condition
				}
				item.Effects = append(item.Effects, newEffect)
			}
		}
	}

	if initial, ok := data["initial"].(int); ok {
		item.Initial = initial
	}

	if reqs, ok := data["reqs"].([]interface{}); ok {
		item.Reqs = make([]string, 0)
		for _, req := range reqs {
			if reqStr, ok := req.(string); ok {
				item.Reqs = append(item.Reqs, reqStr)
			}
		}
	}

	item.Properties = data

	//log.Printf("Item %+v", item)

	return item, nil
}

func (cs *ContentSystem) GetContent(category, name string) (ContentItem, error) {
	if categoryContent, ok := cs.content[category]; ok {
		if item, ok := categoryContent[name]; ok {
			return item, nil
		}
	}
	return ContentItem{}, fmt.Errorf("content not found: %s in category %s", name, category)
}

func (cs *ContentSystem) GetAllContent(category string) map[string]ContentItem {
	return cs.content[category]
}

func (cs *ContentSystem) GetCategories() []string {
	categories := make([]string, 0, len(cs.content))
	for category := range cs.content {
		categories = append(categories, category)
	}
	return categories
}

func (cs *ContentSystem) RegisterPlugin(plugin Plugin) {
	cs.pluginSystem.RegisterPlugin(plugin)
}

func (cs *ContentSystem) CreateCustomContent(contentType string, data map[string]interface{}) (ContentItem, error) {
	return cs.pluginSystem.CreateContent(contentType, data)
}
