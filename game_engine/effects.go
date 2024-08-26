package game_engine

import (
	"fmt"
	"math/rand"
	"strings"
)

type Effect struct {
	Type       string
	Target     string
	Value      float64
	Condition  string
	Expression string
}

type EffectBlock struct {
	TriggerType string
	Effects     []Effect
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
	// Добавьте другие типы эффектов по мере необходимости
	}
}

func (g *Game) applyYieldEffect(player *Player, effect Effect) {
	amount, _ := g.evaluateExpression(effect.Expression, player.getVariables())
	player.AddResource(effect.Target, amount)
}

func (g *Game) applyMultiplyEffect(player *Player, effect Effect) {
	currentAmount := player.State.Resources[effect.Target]
	newAmount, _ := g.evaluateExpression(fmt.Sprintf("%f * %f", currentAmount, effect.Value), nil)
	player.State.Resources[effect.Target] = newAmount
}

func (g *Game) applyGrantEffect(player *Player, effect Effect) {
	amount, _ := g.evaluateExpression(effect.Expression, player.getVariables())
	player.State.Resources[effect.Target] += amount
}

func (g *Game) applySpawnEffect(player *Player, effect Effect) {
	if shiny, ok := g.ContentSystem.Shinies[effect.Target]; ok {
		g.applyShinyEffect(player, shiny)
	}
}

func (g *Game) executeEffectBlock(player *Player, block EffectBlock) {
	for _, effect := range block.Effects {
		if effect.Condition != "" {
			if !g.evaluateCondition(player, effect.Condition) {
				continue
			}
		}
		g.applyEffect(player, effect)
	}
}

func (p *Player) getVariables() map[string]float64 {
	variables := make(map[string]float64)
	for resource, amount := range p.State.Resources {
		variables[resource] = amount
	}
	for building, amount := range p.State.Buildings {
		variables[building] = float64(amount)
	}
	variables["prestige"] = float64(p.State.Prestige)
	return variables
}

func (g *Game) parseEffectString(effectString string) Effect {
	parts := strings.Split(effectString, ":")
	if len(parts) < 2 {
		return Effect{}
	}

	effect := Effect{Type: parts[0], Target: parts[1]}
	if len(parts) > 2 {
		effect.Expression = parts[2]
	}

	return effect
}

func (g *Game) evaluateChance(chanceString string) bool {
	chanceValue, err := g.evaluateExpression(strings.TrimSuffix(chanceString, "%"), nil)
	if err != nil {
		return false
	}
	return rand.Float64()*100 < chanceValue
}