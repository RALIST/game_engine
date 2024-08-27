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

func evaluateExpression(player *Player, expression string) (float64, error) {
	evaluator := NewExpressionEvaluator(player)
	return evaluator.Evaluate(expression)
}

func (g *Game) evaluateCondition(player *Player, condition string) bool {
	result, err := evaluateExpression(player, condition)
	if err != nil {
		log.Printf("Error evaluating condition: %v", err)
		return false
	}
	return result > 0
}

func (g *Game) updatePlayer(player *Player) {
	for id, amount := range player.State.RPS {
		player.State.Items[id].Amount += amount
	}
}

func (g *Game) Buy(player *Player, itemID string) error {
	item := player.State.Items[itemID]
	cost := g.calculateCost(item.Cost, player.GetItemAmount(itemID))

	if player.CanAfford(cost) {
		player.SpendResources(cost)
		item.Amount++
		log.Printf("Player %s bought item: %s (now have %d)", player.ID, item.Name, player.GetItemAmount(itemID))
		player.AddLog(fmt.Sprintf("Bought item: %s (now have %d)", item.Name, player.GetItemAmount(itemID)))

		player.RecalculateState()

		g.EventSystem.Emit("BuildingBought", map[string]interface{}{
			"PlayerID": player.ID,
			"ItemID":   item.ID,
			"Amount":   player.GetItemAmount(itemID),
		})
		return nil
	}

	return fmt.Errorf("cannot afford item: %s", itemID)
}

func (g *Game) Sell(player *Player, itemID string) error {
	item := player.GetItem(itemID)
	if item == nil || item.Amount <= 0 {
		return fmt.Errorf("item not found: %s", itemID)
	}

	g.EventSystem.Emit("BuildingSold", map[string]interface{}{
		"PlayerID": player.ID,
		"ItemID":   itemID,
		"Amount":   item.Amount,
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
