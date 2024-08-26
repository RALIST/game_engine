package game_engine

import (
	"fmt"
	"github.com/ralist/game_engine/game_engine/config"
	"log"
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

func NewGame(cfg *config.GameConfig) (*Game, error) {
	content, err := NewContentSystem(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create content system: %w", err)
	}
	return &Game{
		EventSystem:     NewEventSystem(),
		expressionCache: NewCache(),
		ContentSystem:   content,
		PluginSystem:    NewPluginSystem(),
	}, nil
}

func (g *Game) evaluateExpression(player *Player, expression string) (float64, error) {
	evaluator := NewExpressionEvaluator(player, g)
	return evaluator.Evaluate(expression)
}

func (g *Game) evaluateCondition(player *Player, condition string) bool {
	result, err := g.evaluateExpression(player, condition)
	if err != nil {
		log.Printf("Error evaluating condition: %v", err)
		return false
	}
	return result > 0
}

func (g *Game) updatePlayer(player *Player) {
	totalYield := g.GetTotalYield(player)
	totalMultiplier := g.GetTotalMultiplier(player)
	for name, yield := range totalYield {
		mult := totalMultiplier[name]
		if mult == 0 {
			mult = 1
		}
		amountToAdd := float64(yield) * mult
		player.AddResource(name, amountToAdd)
	}
}

func (g *Game) GetTotalMultiplier(player *Player) map[string]float64 {
	totalMult := make(map[string]float64)
	for name, owned := range player.State.Upgrades {
		if !owned {
			continue
		}

		upgradeItem, err := g.ContentSystem.GetContent("upgrades", name)
		if err != nil {
			log.Printf("Error getting upgrade %s: %v", name, err)
			continue
		}

		for _, effect := range upgradeItem.Effects {
			if effect.Type == "multiply" {
				totalMult[effect.Target] += effect.Value
			}
		}
	}
	return totalMult
}

func (g *Game) GetTotalYield(player *Player) map[string]int {
	totalYield := make(map[string]int)
	for name, amount := range player.State.Buildings {
		if amount == 0 {
			continue
		}
		buildingItem, err := g.ContentSystem.GetContent("buildings", name)
		if err != nil {
			log.Printf("Error getting building %s: %v", name, err)
			continue
		}

		for _, effect := range buildingItem.Effects {
			if effect.Type == "yield" {
				yieldAmount, err := g.evaluateExpression(player, effect.Expression)
				if err != nil {
					log.Printf("Error evaluating yield expression for %s: %v", name, err)
					continue
				}
				totalYield[effect.Target] += int(yieldAmount)
			}
		}
	}
	return totalYield
}

func (g *Game) updateEffects(player *Player, itemType string) error {
	items := player.State.Buildings

	for itemName, amount := range items {
		if amount == 0 {
			continue
		}

		item, err := g.ContentSystem.GetContent(itemType, itemName)
		if err != nil {
			return fmt.Errorf("error getting %s %s: %w", itemType, itemName, err)
		}

		for _, effect := range item.Effects {
			switch effect.Type {
			case "yield":
				if itemType != "buildings" {
					continue
				}

				yieldAmount, err := g.evaluateExpression(player, effect.Expression)
				if err != nil {
					return fmt.Errorf("error evaluating yield expression for %s: %w", itemName, err)
				}
				mult := player.State.Multipliers[effect.Target]
				player.AddResource(effect.Target, yieldAmount*float64(amount)*mult)
			case "multiply":
				currentAmount := player.State.Multipliers[effect.Target]
				newAmount, err := g.evaluateExpression(player, fmt.Sprintf("%f * %f", currentAmount, effect.Value))
				if err != nil {
					return fmt.Errorf("error evaluating upgrade effect for %s: %w", itemName, err)
				}
				player.State.Multipliers[effect.Target] = newAmount
			}
		}
	}
	return nil
}

func (g *Game) updateAchievements(player *Player) error {
	//achievements := g.ContentSystem.GetAllContent("achievements")

	//for _, achievement := range achievements {
	//	if g.evaluateCondition(player, achievement.Condition) {
	//		player.AddAchievement(achievement.Type)
	//		g.applyAchievementRewards(player, achievement)
	//	}
	//}
	return nil
}

func (g *Game) applyAchievementRewards(player *Player, achievement GameItem) {
	for _, effect := range achievement.Effects {
		switch effect.Type {
		case "yield":
			player.AddResource(effect.Target, effect.Value)
		case "multiply":
			player.State.Multipliers[effect.Target] *= effect.Value
		}
	}
}

func (g *Game) applyShinyEffect(player *Player, shinyName string) error {
	shinyItem, err := g.ContentSystem.GetContent("shinies", shinyName)
	if err != nil {
		return fmt.Errorf("error getting shiny %s: %w", shinyName, err)
	}

	for _, effect := range shinyItem.Effects {
		if effect.Type == "yield" {
			player.AddResource(effect.Target, effect.Value)
		}
	}
	return nil
}

func (g *Game) Buy(player *Player, name string) error {
	log.Printf("Player %s is trying to buy %s", player.ID, name)

	item, err := g.ContentSystem.GetContent("upgrades", name)
	if err == nil {
		return g.buyUpgrade(player, item, name)
	}

	item, err = g.ContentSystem.GetContent("buildings", name)
	if err == nil {
		return g.buyBuilding(player, item, name)
	}

	return fmt.Errorf("unknown item to buy: %s", name)
}

func (g *Game) buyUpgrade(player *Player, upgrade GameItem, name string) error {
	owned := player.State.Upgrades[upgrade.Type]
	if owned {
		return fmt.Errorf("upgrade already bought: %s", name)
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
		return nil
	}
	return fmt.Errorf("cannot afford upgrade: %s", name)
}

func (g *Game) buyBuilding(player *Player, building GameItem, name string) error {
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
		return nil
	}
	return fmt.Errorf("cannot afford building: %s", name)
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
		result := amount * float64(owned+1)
		cost[resource] = result
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
		return fmt.Errorf("error getting prestige content: %w", err)
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
