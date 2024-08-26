package game_engine

import (
	"fmt"
)

type Plugin interface {
	Init(game *Game)
	GetContentTypes() []string
	CreateContent(contentType string, data map[string]interface{}) (ContentItem, error)
}

type PluginSystem struct {
	plugins []Plugin
}

func NewPluginSystem() *PluginSystem {
	return &PluginSystem{
		plugins: make([]Plugin, 0),
	}
}

func (ps *PluginSystem) RegisterPlugin(plugin Plugin) {
	ps.plugins = append(ps.plugins, plugin)
}

func (ps *PluginSystem) CreateContent(contentType string, data map[string]interface{}) (ContentItem, error) {
	for _, plugin := range ps.plugins {
		if content, err := plugin.CreateContent(contentType, data); err == nil {
			return content, nil
		}
	}
	return ContentItem{}, fmt.Errorf("no plugin found to handle content type: %s", contentType)
}

func (ps *PluginSystem) InitializePlugins(game *Game) {
	for _, plugin := range ps.plugins {
		plugin.Init(game)
	}
}