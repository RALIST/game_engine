package game_engine

import (
	"fmt"
	"log"
	"time"

	"github.com/ralist/game_engine/config"
)

// Interfaces
type UIInterface interface {
	DisplayGameState(state GameState)
	GetUserInput() string
}

type DatabaseInterface interface {
	SaveGame(data []byte) error
	LoadGame() ([]byte, error)
	SavePlayer(playerID string, data []byte) error
	LoadPlayer(playerID string) ([]byte, error)
}

type CommandSystemInterface interface {
	ExecuteCommand(player *Player, commandName string, args []string) error
	GetCommandList() string
}

// Structs
type GameState struct {
	Players map[string]*Player `json:"players"`
	Config  *config.GameConfig `json:"config"`
}

type Game struct {
	State           GameState
	CommandSystem   CommandSystemInterface
	EventSystem     *EventSystem
	expressionCache *Cache
	ContentSystem   *ContentSystem
}

type EventSystem struct {
	listeners map[string][]func(map[string]interface{})
}

// Game methods
func NewGame(cfg *config.GameConfig) *Game {
	return &Game{
		State: GameState{
			Players: make(map[string]*Player),
			Config:  cfg,
		},
		EventSystem:     NewEventSystem(),
		expressionCache: NewCache(),
		ContentSystem: NewContentSystem(cfg),
    }
	}
}

func (g *Game) evaluateExpression(expression string, variables map[string]float64) (float64, error) {
	cacheKey := fmt.Sprintf("%s_%v", expression, variables)
	if cachedResult, found := g.expressionCache.Get(cacheKey); found {
		return cachedResult.(float64), nil
	}

	ee := NewExpressionEvaluator()
	for name, value := range variables {
		ee.SetVariable(name, value)
	}
	result, err := ee.Evaluate(expression)
	if err == nil {
		g.expressionCache.Set(cacheKey, result, 5*time.Minute) // Cache for 5 minutes
	}
	return result, err
}

func (g *Game) evaluateCondition(player *Player, condition string) bool {
	variables := make(map[string]float64)
	for resource, amount := range player.State.Resources {
		variables[resource] = amount
	}
	for building, amount := range player.State.Buildings {
		variables[building] = float64(amount)
	}
	variables["prestige"] = float64(player.State.Prestige)

	result, err := g.evaluateExpression(condition, variables)
	if err != nil {
		log.Printf("Error evaluating condition: %v", err)
		return false
	}
	return result > 0
}

func (g *Game) updateBuildings(player *Player) {
	for buildingName, amount := range player.State.Buildings {
		buildingConfig := g.State.Config.Buildings[buildingName]
		for _, effect := range buildingConfig.Effects {
			if effect.Type == "yield" {
				variables := map[string]float64{
					"tier": float64(amount),
				}
				yieldAmount, err := g.evaluateExpression(effect.Expression, variables)
				if err != nil {
					log.Printf("Error evaluating yield expression for %s: %v", buildingName, err)
					continue
				}
				player.AddResource(effect.Target, yieldAmount*float64(amount))
				log.Printf("Player %s: %s yielded %.2f %s (now have %.2f)",
					player.ID, buildingName, yieldAmount, effect.Target, player.State.Resources[effect.Target])
			}
		}
	}
}

func (g *Game) updateUpgrades(player *Player) {
	for upgradeName, owned := range player.State.Upgrades {
		if owned {
			upgradeConfig := g.State.Config.Upgrades[upgradeName]
			for _, effect := range upgradeConfig.Effects {
				if effect.Type == "multiply" {
					currentAmount := player.State.Resources[effect.Target]
					newAmount, err := g.evaluateExpression(fmt.Sprintf("%f * %f", currentAmount, effect.Value), nil)
					if err != nil {
						log.Printf("Error evaluating upgrade effect for %s: %v", upgradeName, err)
						continue
					}
					player.State.Resources[effect.Target] = newAmount
					log.Printf("Player %s: %s multiplied %s by %.2f (now have %.2f)",
						player.ID, upgradeName, effect.Target, effect.Value, newAmount)
				}
			}
		}
	}
}

func (g *Game) applyShinyEffect(player *Player, shiny config.Shiny) {
	for _, effect := range shiny.Effects {
		if effect.Type == "yield" {
			player.AddResource(effect.Target, effect.Value)
			log.Printf("Player %s: Shiny %s yielded %.2f %s (now have %.2f)",
				player.ID, shiny.Name, effect.Value, effect.Target, player.State.Resources[effect.Target])
		}
	}
}

func (g *Game) calculateCost(baseCost map[string]float64, owned int) map[string]float64 {
	cost := make(map[string]float64)
	for resource, amount := range baseCost {
		expression := fmt.Sprintf("%f * %d", amount, owned+1)
		result, err := g.evaluateExpression(expression, nil)
		if err != nil {
			log.Printf("Error calculating cost for resource %s: %v", resource, err)
			cost[resource] = amount * float64(owned+1) // Fallback to simple multiplication
		} else {
			cost[resource] = result
		}
	}
	return cost
}

func (g *Game) PerformPrestige(player *Player) error {
	prestigeConfig := g.State.Config.Prestige
	if !player.CanAfford(prestigeConfig.Cost) {
		return fmt.Errorf("cannot afford prestige cost")
	}

	player.SpendResources(prestigeConfig.Cost)
	player.ResetProgress()

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

	g.EventSystem.Emit("BuildingSold", map[string]interface{}{
		"PlayerID":     player.ID,
		"BuildingName": name,
		"Amount":       player.State.Buildings[name],
	})

	return nil
}

func (g *Game) calculateSellPrice(baseCost map[string]float64) map[string]float64 {
	sellPrice := make(map[string]float64)
	for resource, amount := range baseCost {
		sellPrice[resource] = amount * 0.5
	}
	return sellPrice
}

func (g *Game) updatePlayer(player *Player) {
	g.updateBuildings(player)
	g.updateUpgrades(player)
	g.updateAchievements(player)
}

func (g *Game) updateAchievements(player *Player) {
	for _, achievement := range g.State.Config.Achievements {
		currentLevel := player.GetAchievementLevel(achievement.Name)
		for _, level := range achievement.Levels {
			if level.Level > currentLevel && g.evaluateCondition(player, level.Condition) {
				player.SetAchievementLevel(achievement.Name, level.Level)
				player.AddLog(fmt.Sprintf("Achievement unlocked: %s (Level %d)", achievement.Name, level.Level))

				for resource, amount := range level.Rewards {
					player.AddResource(resource, amount)
				}

				g.EventSystem.Emit("AchievementUnlocked", map[string]interface{}{
					"PlayerID":        player.ID,
					"AchievementName": achievement.Name,
					"Level":           level.Level,
				})
			}
		}
	}
}

func (g *Game) Buy(player *Player, name string) {
	log.Printf("Player %s is trying to buy %s", player.ID, name)

	if upgrade, ok := g.State.Config.Upgrades[name]; ok {
		g.buyUpgrade(player, upgrade, name)
		return
	}

	if building, ok := g.State.Config.Buildings[name]; ok {
		g.buyBuilding(player, building, name)
		return
	}

	log.Printf("Player %s tried to buy unknown item: %s", player.ID, name)
	player.AddLog("Unknown item to buy")
}

func (g *Game) buyUpgrade(player *Player, upgrade config.Upgrade, name string) {
	if player.CanAfford(upgrade.Cost) {
		player.SpendResources(upgrade.Cost)
		player.AddUpgrade(upgrade.Name)
		log.Printf("Player %s bought upgrade: %s", player.ID, upgrade.Name)
		player.AddLog(fmt.Sprintf("Bought upgrade: %s", upgrade.Name))

		g.EventSystem.Emit("UpgradeBought", map[string]interface{}{
			"PlayerID":    player.ID,
			"UpgradeName": upgrade.Name,
		})
	} else {
		log.Printf("Player %s cannot afford upgrade: %s", player.ID, name)
		player.AddLog("Cannot afford this upgrade")
	}
}

func (g *Game) buyBuilding(player *Player, building config.Building, name string) {
	cost := g.calculateCost(building.Cost, player.GetBuildingAmount(name))
	if player.CanAfford(cost) {
		player.SpendResources(cost)
		player.AddBuilding(building.Name)
		log.Printf("Player %s bought building: %s (now have %d)", player.ID, building.Name, player.GetBuildingAmount(name))
		player.AddLog(fmt.Sprintf("Bought building: %s (now have %d)", building.Name, player.GetBuildingAmount(name)))

		g.EventSystem.Emit("BuildingBought", map[string]interface{}{
			"PlayerID":     player.ID,
			"BuildingName": building.Name,
			"Amount":       player.GetBuildingAmount(name),
		})
	} else {
		log.Printf("Player %s cannot afford building: %s", player.ID, name)
		player.AddLog("Cannot afford this building")
	}
}

// GameEngine methods
func NewGameEngine(cfg *config.GameConfig, ui UIInterface, db DatabaseInterface) *GameEngine {
	game := NewGame(cfg)
	engine := &GameEngine{
		game: game,
		ui:   ui,
		db:   db,
	}
	return engine
}

func (ge *GameEngine) updatePlayer(player *Player) {
	ge.game.updatePlayer(player)
}

// EventSystem methods
func NewEventSystem() *EventSystem {
	return &EventSystem{
		listeners: make(map[string][]func(map[string]interface{})),
	}
}

func (es *EventSystem) On(eventName string, listener func(map[string]interface{})) {
	es.listeners[eventName] = append(es.listeners[eventName], listener)
}

func (es *EventSystem) Emit(eventName string, data map[string]interface{}) {
	if listeners, ok := es.listeners[eventName]; ok {
		for _, listener := range listeners {
			listener(data)
		}
	}
}
