package main

import (
	"fmt"
	"log"

	"github.com/yourusername/idle-game-engine/config"
)

// evaluateCondition оценивает условие достижения для конкретного игрока
func (g *Game) evaluateCondition(player *Player, condition string) bool {
	ee := NewExpressionEvaluator()

	// Set variables based on player state
	for resource, amount := range player.State.Resources {
		ee.SetVariable(resource, amount)
	}
	for building, amount := range player.State.Buildings {
		ee.SetVariable(building, float64(amount))
	}
	ee.SetVariable("prestige", float64(player.State.Prestige))

	result, err := ee.Evaluate(condition)
	if err != nil {
		log.Printf("Error evaluating condition: %v", err)
		return false
	}
	return result > 0
}

func (g *Game) PerformPrestige(player *Player) error {
	prestigeConfig := g.State.Config.Prestige
	if player.CanAfford(prestigeConfig.Cost) {
		player.SpendResources(prestigeConfig.Cost)
		player.ResetProgress()

		// Apply prestige effects
		for _, effect := range prestigeConfig.Effects {
			switch effect.Type {
			case "multiply":
				currentAmount := player.State.Resources[effect.Target]
				player.State.Resources[effect.Target] = currentAmount * effect.Value
			case "reset":
				if effect.Target == "all" {
					for resource := range player.State.Resources {
						player.State.Resources[resource] = g.State.Config.Resources[resource].Initial
					}
				} else {
					player.State.Resources[effect.Target] = g.State.Config.Resources[effect.Target].Initial
				}
			}
		}

		player.AddLog(fmt.Sprintf("Performed prestige: %s", prestigeConfig.Name))

		g.EventSystem.Emit("Prestige", map[string]interface{}{
			"PlayerID":      player.ID,
			"PrestigeLevel": player.State.Prestige,
		})

		return nil
	}
	return fmt.Errorf("cannot afford prestige cost")
}

// UIInterface определяет методы для взаимодействия с пользовательским интерфейсом
type UIInterface interface {
	DisplayGameState(state GameState)
	GetUserInput() string
}

// DatabaseInterface определяет методы для взаимодействия с базой данных
type DatabaseInterface interface {
	SaveGame(data []byte) error
	LoadGame() ([]byte, error)
	SavePlayer(playerID string, data []byte) error
	LoadPlayer(playerID string) ([]byte, error)
}

// CommandSystemInterface определяет методы для системы команд
type CommandSystemInterface interface {
	ExecuteCommand(player *Player, commandName string, args []string) error
	GetCommandList() string
}

// GameState представляет общее состояние игры
type GameState struct {
	Players map[string]*Player `json:"players"`
	Config  *config.GameConfig `json:"config"`
}

// Game представляет основную структуру игры
type Game struct {
	State         GameState
	CommandSystem CommandSystemInterface
	EventSystem   *EventSystem
}

// GameEngine представляет основной движок игры
type GameEngine struct {
	game *Game
	ui   UIInterface
	db   DatabaseInterface
}

// NewGameEngine создает новый экземпляр GameEngine
func NewGameEngine(cfg *config.GameConfig, ui UIInterface, db DatabaseInterface) *GameEngine {
	game := NewGame(cfg)
	engine := &GameEngine{
		game: game,
		ui:   ui,
		db:   db,
	}
	return engine
}

// NewGame создает новую игру
func NewGame(cfg *config.GameConfig) *Game {
	return &Game{
		State: GameState{
			Players: make(map[string]*Player),
			Config:  cfg,
		},
		EventSystem: NewEventSystem(),
	}
}

// Sell продает здание игрока
func (g *Game) Sell(player *Player, name string) error {
	buildingAmount := player.GetBuildingAmount(name)
	if buildingAmount <= 0 {
		return fmt.Errorf("no %s buildings to sell", name)
	}

	buildingConfig, ok := g.State.Config.Buildings[name]
	if !ok {
		return fmt.Errorf("unknown building: %s", name)
	}

	sellPrice := g.calculateSellPrice(buildingConfig.Cost)
	for resource, amount := range sellPrice {
		player.AddResource(resource, amount)
	}

	player.State.Buildings[name]--
	player.AddLog(fmt.Sprintf("Sold building: %s (now have %d)", name, player.State.Buildings[name]))

	// Генерируем событие продажи здания
	g.EventSystem.Emit("BuildingSold", map[string]interface{}{
		"PlayerID":     player.ID,
		"BuildingName": name,
		"Amount":       player.State.Buildings[name],
	})

	return nil
}

// calculateSellPrice вычисляет цену продажи здания
func (g *Game) calculateSellPrice(baseCost map[string]float64) map[string]float64 {
	sellPrice := make(map[string]float64)
	for resource, amount := range baseCost {
		sellPrice[resource] = amount * 0.5 // Продажа за 50% от базовой стоимости
	}
	return sellPrice
}

func (ge *GameEngine) updatePlayer(player *Player) {
	ge.game.updatePlayer(player)
}

// updatePlayer обновляет состояние игры для одного игрока
func (ge *Game) updatePlayer(player *Player) {
	// Обновление ресурсов на основе зданий
	for buildingName, amount := range player.State.Buildings {
		buildingConfig := ge.State.Config.Buildings[buildingName]
		for _, effect := range buildingConfig.Effects {
			if effect.Type == "yield" {
				yieldAmount := effect.Value * float64(amount)
				player.AddResource(effect.Target, yieldAmount)
				log.Printf("Player %s: %s yielded %.2f %s (now have %.2f)",
					player.ID, buildingName, yieldAmount, effect.Target, player.State.Resources[effect.Target])
			}
		}
	}

	// Обновление ресурсов на основе улучшений
	for upgradeName, owned := range player.State.Upgrades {
		if owned {
			upgradeConfig := ge.State.Config.Upgrades[upgradeName]
			for _, effect := range upgradeConfig.Effects {
				if effect.Type == "multiply" {
					currentAmount := player.State.Resources[effect.Target]
					player.State.Resources[effect.Target] = currentAmount * effect.Value
					log.Printf("Player %s: %s multiplied %s by %.2f (now have %.2f)",
						player.ID, upgradeName, effect.Target, effect.Value, player.State.Resources[effect.Target])
				}
			}
		}
	}

	// Обновление достижений
	for _, achievement := range ge.State.Config.Achievements {
		if !player.GetAchievementStatus(achievement.Name) && ge.evaluateCondition(player, achievement.Condition) {
			player.SetAchievement(achievement.Name)
			player.AddLog("Achievement unlocked: " + achievement.Name)

			// Генерируем событие достижения
			ge.EventSystem.Emit("AchievementUnlocked", map[string]interface{}{
				"PlayerID":        player.ID,
				"AchievementName": achievement.Name,
			})
		}
	}
}

// Buy покупает здание или улучшение для конкретного игрока
func (g *Game) Buy(player *Player, name string) {
	log.Printf("Player %s is trying to buy %s", player.ID, name)
	if upgrade, ok := g.State.Config.Upgrades[name]; ok {
		if player.CanAfford(upgrade.Cost) {
			player.SpendResources(upgrade.Cost)
			player.AddUpgrade(upgrade.Name)
			log.Printf("Player %s bought upgrade: %s", player.ID, upgrade.Name)
			player.AddLog(fmt.Sprintf("Bought upgrade: %s", upgrade.Name))

			// Генерируем событие покупки улучшения
			g.EventSystem.Emit("UpgradeBought", map[string]interface{}{
				"PlayerID":    player.ID,
				"UpgradeName": upgrade.Name,
			})
		} else {
			log.Printf("Player %s cannot afford upgrade: %s", player.ID, name)
			player.AddLog("Cannot afford this upgrade")
		}
		return
	}

	if building, ok := g.State.Config.Buildings[name]; ok {
		cost := g.calculateCost(building.Cost, player.GetBuildingAmount(name))
		if player.CanAfford(cost) {
			player.SpendResources(cost)
			player.AddBuilding(building.Name)
			log.Printf("Player %s bought building: %s (now have %d)", player.ID, building.Name, player.GetBuildingAmount(name))
			player.AddLog(fmt.Sprintf("Bought building: %s (now have %d)", building.Name, player.GetBuildingAmount(name)))

			// Генерируем событие покупки здания
			g.EventSystem.Emit("BuildingBought", map[string]interface{}{
				"PlayerID":     player.ID,
				"BuildingName": building.Name,
				"Amount":       player.GetBuildingAmount(name),
			})
		} else {
			log.Printf("Player %s cannot afford building: %s", player.ID, name)
			player.AddLog("Cannot afford this building")
		}
		return
	}

	log.Printf("Player %s tried to buy unknown item: %s", player.ID, name)
	player.AddLog("Unknown item to buy")
}

// calculateCost вычисляет стоимость здания с учетом уже купленных
func (g *Game) calculateCost(baseCost map[string]float64, owned int) map[string]float64 {
	cost := make(map[string]float64)
	for resource, amount := range baseCost {
		cost[resource] = amount * float64(owned+1)
	}
	return cost
}

// EventSystem представляет систему событий в игре
type EventSystem struct {
	listeners map[string][]func(map[string]interface{})
}

// NewEventSystem создает новую систему событий
func NewEventSystem() *EventSystem {
	return &EventSystem{
		listeners: make(map[string][]func(map[string]interface{})),
	}
}

// On регистрирует новый слушатель для события
func (es *EventSystem) On(eventName string, listener func(map[string]interface{})) {
	es.listeners[eventName] = append(es.listeners[eventName], listener)
}

// Emit генерирует событие
func (es *EventSystem) Emit(eventName string, data map[string]interface{}) {
	if listeners, ok := es.listeners[eventName]; ok {
		for _, listener := range listeners {
			listener(data)
		}
	}
}

func (ge *GameEngine) updateAchievements(player *Player) {
	for _, achievement := range ge.game.State.Config.Achievements {
		currentLevel := player.GetAchievementLevel(achievement.Name)
		for _, level := range achievement.Levels {
			if level.Level > currentLevel && ge.game.evaluateCondition(player, level.Condition) {
				player.SetAchievementLevel(achievement.Name, level.Level)
				player.AddLog(fmt.Sprintf("Achievement unlocked: %s (Level %d)", achievement.Name, level.Level))

				// Apply rewards
				for resource, amount := range level.Rewards {
					player.AddResource(resource, amount)
				}

				ge.game.EventSystem.Emit("AchievementUnlocked", map[string]interface{}{
					"PlayerID":        player.ID,
					"AchievementName": achievement.Name,
					"Level":           level.Level,
				})
			}
		}
	}
}
