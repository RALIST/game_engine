package game_engine

import (
	"fmt"
	"github.com/ralist/game_engine/game_engine/config"
)

// GameItem представляет собой элемент игрового контента
type GameItem struct {
	Type        string                 `yaml:"type"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Cost        map[string]float64     `yaml:"cost"`
	Effects     []config.Effect        `yaml:"effects"`
	Initial     int                    `yaml:"initial"`
	Reqs        []string               `yaml:"reqs"`
	Properties  map[string]interface{} `yaml:"properties"`
}

// ContentSystem управляет всем игровым контентом
type ContentSystem struct {
	content           map[string]map[string]GameItem
	effectDefinitions map[string]config.EffectDefinition
	pluginSystem      *PluginSystem
}

// NewContentSystem создает новую систему контента на основе конфигурации игры
func NewContentSystem(cfg *config.GameConfig) (*ContentSystem, error) {
	cs := &ContentSystem{
		content:           make(map[string]map[string]GameItem),
		effectDefinitions: cfg.EffectDefinitions,
		pluginSystem:      NewPluginSystem(),
	}

	if err := cs.parseContent(cfg.Content); err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return cs, nil
}

// parseContent парсит контент из конфигурации
func (cs *ContentSystem) parseContent(content map[string]map[string]interface{}) error {
	for category, items := range content {
		cs.content[category] = make(map[string]GameItem)
		for name, data := range items {
			convertedData, ok := data.(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf("invalid data format for item %s in category %s", name, category)
			}
			item, err := cs.createContentItem(name, convertMapInterfaceToMapString(convertedData))
			if err != nil {
				return fmt.Errorf("failed to create content item %s in category %s: %w", name, category, err)
			}
			cs.content[category][name] = item
		}
	}
	return nil
}

// convertMapInterfaceToMapString преобразует map[interface{}]interface{} в map[string]interface{}
func convertMapInterfaceToMapString(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		key, ok := k.(string)
		if !ok {
			continue // Пропускаем ключи, которые не являются строками
		}
		switch value := v.(type) {
		case map[interface{}]interface{}:
			result[key] = convertMapInterfaceToMapString(value)
		case []interface{}:
			result[key] = convertSliceInterfaceToSliceString(value)
		default:
			result[key] = v
		}
	}
	return result
}

// convertSliceInterfaceToSliceString преобразует []interface{} рекурсивно
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

// createContentItem создает элемент контента из данных
func (cs *ContentSystem) createContentItem(name string, data map[string]interface{}) (GameItem, error) {
	item := GameItem{
		Type:       name,
		Properties: data,
	}

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
		item.Cost = make(map[string]float64, len(cost))
		for resource, amount := range cost {
			if floatAmount, ok := amount.(float64); ok {
				item.Cost[resource] = floatAmount
			} else if intAmount, ok := amount.(int); ok {
				item.Cost[resource] = float64(intAmount)
			} else {
				return GameItem{}, fmt.Errorf("invalid cost value for resource %s", resource)
			}
		}
	}

	if effects, ok := data["effects"].([]interface{}); ok {
		item.Effects = make([]config.Effect, 0, len(effects))
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
				} else if intValue, ok := effectMap["value"].(int); ok {
					newEffect.Value = float64(intValue)
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
		item.Reqs = make([]string, 0, len(reqs))
		for _, req := range reqs {
			if reqStr, ok := req.(string); ok {
				item.Reqs = append(item.Reqs, reqStr)
			}
		}
	}

	return item, nil
}

// GetContent возвращает элемент контента по категории и имени
func (cs *ContentSystem) GetContent(category, name string) (GameItem, error) {
	categoryContent, ok := cs.content[category]
	if !ok {
		return GameItem{}, fmt.Errorf("category not found: %s", category)
	}
	item, ok := categoryContent[name]
	if !ok {
		return GameItem{}, fmt.Errorf("content not found: %s in category %s", name, category)
	}
	return item, nil
}

// GetAllContent возвращает все элементы контента в указанной категории
func (cs *ContentSystem) GetAllContent(category string) map[string]GameItem {
	return cs.content[category]
}

// GetCategories возвращает список всех категорий контента
func (cs *ContentSystem) GetCategories() []string {
	categories := make([]string, 0, len(cs.content))
	for category := range cs.content {
		categories = append(categories, category)
	}
	return categories
}

// RegisterPlugin регистрирует новый плагин в системе
func (cs *ContentSystem) RegisterPlugin(plugin Plugin) {
	cs.pluginSystem.RegisterPlugin(plugin)
}

// CreateCustomContent создает пользовательский контент с помощью плагинов
func (cs *ContentSystem) CreateCustomContent(contentType string, data map[string]interface{}) (GameItem, error) {
	return cs.pluginSystem.CreateContent(contentType, data)
}
