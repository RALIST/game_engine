package game_engine

import (
	"fmt"
	"github.com/ralist/game_engine/game_engine/config"
	"log"
	"time"
)

type UIInterface interface {
	DisplayGameState(state GameState)
	GetUserInput() string
}

type DatabaseInterface interface {
	LoadPlayers() ([]interface{}, error)
	SavePlayer(playerID string, data []byte) error
	LoadPlayer(playerID string) ([]byte, error)
}

type CommandSystemInterface interface {
	ExecuteCommand(player *Player, commandName string, args []string) error
	GetCommandList() string
}

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
	PluginSystem    *PluginSystem
}

type AchievementLevel struct {
	Level     int                `yaml:"level"`
	Condition string             `yaml:"condition"`
	Rewards   map[string]float64 `yaml:"rewards"`
}

func NewGame(cfg *config.GameConfig) *Game {
	content, err := NewContentSystem(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &Game{
		State: GameState{
			Players: make(map[string]*Player),
			Config:  cfg,
		},
		EventSystem:     NewEventSystem(),
		expressionCache: NewCache(),
		ContentSystem:   content,
		PluginSystem:    NewPluginSystem(),
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
	variables := player.getVariables()
	result, err := g.evaluateExpression(condition, variables)
	if err != nil {
		log.Printf("Error evaluating condition: %v", err)
		return false
	}
	return result > 0
}

func (g *Game) updatePlayer(player *Player) {
	totalYield := g.GetTotalYield(player)
	totalMultiplier := g.GetTotalMultiplier(player)
	for name, _ := range player.State.Resources {
		mult := totalMultiplier[name]
		if mult == 0 {
			mult = 1
		}
		amountToAdd := float64(totalYield[name]) * mult
		player.AddResource(name, amountToAdd)
	}
}

func (g *Game) GetTotalMultiplier(player *Player) map[string]float64 {
	totalMult := map[string]float64{}
	for name, owned := range player.State.Upgrades {
		if !owned {
			return nil
		}

		buildingItem, err := g.ContentSystem.GetContent("upgrades", name)
		if err != nil {
			log.Printf("Error getting building %s: %v", name, err)
			continue
		}

		for _, effect := range buildingItem.Effects {
			if effect.Type == "multiply" {
				value := effect.Value
				totalMult[effect.Target] += value
			}
		}
	}
	return totalMult
}

func (g *Game) GetTotalYield(player *Player) map[string]int {
	totalYield := map[string]int{}
	for name, amount := range player.State.Buildings {
		if amount == 0 {
			continue
		}
		buildingItem, err := g.ContentSystem.GetContent("buildings", name)
		if err != nil {
			log.Printf("Error getting amount %d: %v", amount, err)
			continue
		}

		for _, effect := range buildingItem.Effects {
			if effect.Type == "yield" {
				variables := map[string]float64{
					"count": float64(amount),
				}

				yieldAmount, err := g.evaluateExpression(effect.Expression, variables)
				if err != nil {
					log.Printf("Error evaluating yield expression for %d: %v", amount, err)
					continue
				}
				totalYield[effect.Target] += int(yieldAmount)
				log.Printf("Player %s: %s yielded %.2f %s (for %d buildings)",
					player.ID, buildingItem.Type, yieldAmount, effect.Target, amount)
			}
		}
	}
	log.Println("Total yield:", totalYield)

	return totalYield
}

func (g *Game) updateBuildings(player *Player) {
	for buildingName, amount := range player.State.Buildings {
		if amount == 0 {
			continue
		}

		buildingItem, err := g.ContentSystem.GetContent("buildings", buildingName)
		if err != nil {
			log.Printf("Error getting building %s: %v", buildingName, err)
			continue
		}

		for _, effect := range buildingItem.Effects {
			if effect.Type == "yield" {
				variables := map[string]float64{
					"tier": float64(amount),
				}
				yieldAmount, err := g.evaluateExpression(effect.Expression, variables)
				if err != nil {
					log.Printf("Error evaluating yield expression for %s: %v", buildingName, err)
					continue
				}
				mult := player.State.Multipliers[effect.Target]

				player.AddResource(effect.Target, yieldAmount*float64(amount)*mult)
				log.Printf("Player %s: %s yielded %.2f %s (now have %d)",
					player.ID, buildingName, yieldAmount, effect.Target, player.State.Resources[effect.Target])
			}
		}
	}
}

func (g *Game) updateUpgrades(player *Player) {
	for upgradeName, owned := range player.State.Upgrades {
		if owned {
			upgradeItem, err := g.ContentSystem.GetContent("upgrades", upgradeName)
			if err != nil {
				log.Printf("Error getting upgrade %s: %v", upgradeName, err)
				continue
			}

			for _, effect := range upgradeItem.Effects {
				if effect.Type == "multiply" {
					currentAmount := player.State.Multipliers[effect.Target]
					newAmount, err := g.evaluateExpression(fmt.Sprintf("%f * %f", currentAmount, effect.Value), nil)
					if err != nil {
						log.Printf("Error evaluating upgrade effect for %s: %v", upgradeName, err)
						continue
					}
					player.State.Multipliers[effect.Target] = newAmount
					log.Printf("Player %s: %s multiplied %s by %.2f (now have %.2f)",
						player.ID, upgradeName, effect.Target, effect.Value, newAmount)
				}
			}
		}
	}
}

func (g *Game) updateAchievements(player *Player) {
	g.ContentSystem.GetAllContent("achievements")
}

func (g *Game) applyShinyEffect(player *Player, shinyName string) {
	shinyItem, err := g.ContentSystem.GetContent("shinies", shinyName)
	if err != nil {
		log.Printf("Error getting shiny %s: %v", shinyName, err)
		return
	}

	for _, effect := range shinyItem.Effects {
		if effect.Type == "yield" {
			player.AddResource(effect.Target, effect.Value)
			log.Printf("Player %s: Shiny %s yielded %.2f %s (now have %.2f)",
				player.ID, shinyItem.Name, effect.Value, effect.Target, player.State.Resources[effect.Target])
		}
	}
}

func (g *Game) Buy(player *Player, name string) {
	log.Printf("Player %s is trying to buy %s", player.ID, name)

	item, err := g.ContentSystem.GetContent("upgrades", name)
	if err == nil {
		g.buyUpgrade(player, item, name)
		return
	}

	item, err = g.ContentSystem.GetContent("buildings", name)
	if err == nil {
		g.buyBuilding(player, item, name)
		return
	}

	log.Printf("Player %s tried to buy unknown item: %s", player.ID, name)
	player.AddLog("Unknown item to buy")
}

func (g *Game) buyUpgrade(player *Player, upgrade ContentItem, name string) {
	owned := player.State.Upgrades[upgrade.Type]
	if owned {
		log.Printf("Upgrade already bought %s %s", player.ID, name)
		return
	}

	if player.CanAfford(upgrade.Cost) {
		player.SpendResources(upgrade.Cost)
		player.AddUpgrade(upgrade.Type)
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

func (g *Game) buyBuilding(player *Player, building ContentItem, name string) {
	cost := g.calculateCost(building.Cost, player.GetBuildingAmount(name))
	if player.CanAfford(cost) {
		player.SpendResources(cost)
		player.AddBuilding(name)
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

func (g *Game) Sell(player *Player, name string) error {
	buildingAmount := player.GetBuildingAmount(name)
	if buildingAmount <= 0 {
		return fmt.Errorf("no %s buildings to sell", name)
	}

	buildingItem, err := g.ContentSystem.GetContent("buildings", name)
	if err != nil {
		return fmt.Errorf("unknown building: %s", name)
	}

	sellPrice := g.calculateSellPrice(buildingItem.Cost)
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

func (g *Game) calculateCost(baseCost map[string]float64, owned int) map[string]float64 {
	cost := make(map[string]float64)
	for resource, amount := range baseCost {
		expression := fmt.Sprintf("%f * %d", amount, owned+1)
		result, err := g.evaluateExpression(expression, nil)
		if err != nil {
			log.Printf("Error calculating cost for resource %s: %v", resource, err)
			cost[resource] = amount * float64((owned + 1)) // Fallback to simple multiplication
		} else {
			cost[resource] = result
		}
	}
	return cost
}

func (g *Game) calculateSellPrice(baseCost map[string]float64) map[string]float64 {
	sellPrice := make(map[string]float64)
	for resource, amount := range baseCost {
		sellPrice[resource] = amount * 0.5
	}
	return sellPrice
}

func (g *Game) PerformPrestige(player *Player) error {
	prestigeItem, err := g.ContentSystem.GetContent("prestige", "prestige")
	if err != nil {
		return fmt.Errorf("error getting prestige content: %v", err)
	}

	if !player.CanAfford(prestigeItem.Cost) {
		return fmt.Errorf("cannot afford prestige cost")
	}

	player.SpendResources(prestigeItem.Cost)
	player.ResetProgress()

	for _, effect := range prestigeItem.Effects {
		switch effect.Type {
		case "multiply":
			currentAmount := player.State.Resources[effect.Target]
			player.State.Resources[effect.Target] = uint64(float64(currentAmount) * effect.Value)
		case "reset":
			if effect.Target == "all" {
				for resource := range player.State.Resources {
					resourceItem, err := g.ContentSystem.GetContent("resources", resource)
					if err != nil {
						log.Printf("Error getting resource %s: %v", resource, err)
						continue
					}
					player.State.Resources[resource] = uint64(resourceItem.Initial)
				}
			} else {
				resourceItem, err := g.ContentSystem.GetContent("resources", effect.Target)
				if err != nil {
					log.Printf("Error getting resource %s: %v", effect.Target, err)
				} else {
					player.State.Resources[effect.Target] = uint64(resourceItem.Initial)
				}
			}
		}
	}

	player.AddLog(fmt.Sprintf("Performed prestige: %s", prestigeItem.Name))
	g.EventSystem.Emit("Prestige", map[string]interface{}{
		"PlayerID":      player.ID,
		"PrestigeLevel": player.State.Prestige,
	})

	return nil
}
