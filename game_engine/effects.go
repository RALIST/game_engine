package game_engine

import (
	"log"
	"math/rand"
)

type Effect struct {
	Type       string  `yaml:"type"`
	Target     string  `yaml:"target"`
	Value      float64 `yaml:"value"`
	Expression string  `yaml:"expression"`
	Condition  string  `yaml:"condition"`
}

type EffectBlock struct {
	Effects []Effect
	Chance  string
}

func (g *Game) applyEffect(player *Player, effect Effect) {
	switch effect.Type {
	case "yield":
		g.applyYieldEffect(player, effect)
	case "multiply":
		g.applyMultiplyEffect(player, effect)
	case "grant":
		g.applyGrantEffect(player, effect)
	case "spawn":
		g.applySpawnEffect(player, effect)
	default:
		log.Printf("Unknown effect type: %s", effect.Type)
	}
}

func (g *Game) applyYieldEffect(player *Player, effect Effect) {
	amount, err := evaluateExpression(player, effect.Expression)
	if err != nil {
		log.Printf("Error evaluating yield expression: %v", err)
		return
	}
	player.AddItem(effect.Target, amount)
}

func (g *Game) applyMultiplyEffect(player *Player, effect Effect) {
	currentAmount := player.State.Resources[effect.Target]
	newAmount := float64(currentAmount) * effect.Value
	player.State.Resources[effect.Target] = uint64(newAmount)
}

func (g *Game) applyGrantEffect(player *Player, effect Effect) {
	player.SetShinyState(effect.Target, ShinyState{Active: true, LastSpawn: player.State.LastSaveTime})
}

func (g *Game) applySpawnEffect(player *Player, effect Effect) {
	// Implement spawn effect logic here
}

func (g *Game) executeEffectBlock(player *Player, block EffectBlock) {
	if block.Chance != "" {
		if !g.evaluateChance(block.Chance) {
			return
		}
	}

	for _, effect := range block.Effects {
		g.applyEffect(player, effect)
	}
}

func (g *Game) evaluateChance(chanceString string) bool {
	//chance, err := g.evaluateExpression(chanceString, nil)
	//if err != nil {
	//	log.Printf("Error evaluating chance expression: %v", err)
	//	return false
	//}
	return rand.Float64() < 1
}

func (g *Game) parseEffectString(effectString string) Effect {
	// Implement parsing logic for effect strings
	// This is a placeholder implementation
	return Effect{
		Type:   "yield",
		Target: "coins",
		Value:  1,
	}
}
