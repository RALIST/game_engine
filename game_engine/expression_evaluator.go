package game_engine

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Knetic/govaluate"
)

type ExpressionEvaluator struct {
	player *Player
	game   *Game
}

func NewExpressionEvaluator(player *Player, game *Game) *ExpressionEvaluator {
	return &ExpressionEvaluator{
		player: player,
		game:   game,
	}
}

func (ee *ExpressionEvaluator) Evaluate(expression string) (float64, error) {
	functions := map[string]govaluate.ExpressionFunction{
		"have":    ee.have(ee.player),
		"no":      ee.no(ee.player),
		"random":  ee.random,
		"frandom": ee.frandom,
		"chance":  ee.chance,
		"max":     ee.max,
		"min":     ee.min,
		"floor":   ee.floor,
		"ceil":    ee.ceil,
		"round":   ee.round,
		"roundr":  ee.roundr,
		"pow":     ee.pow,
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %w", err)
	}

	params := ee.getParameters(ee.player, ee.game)
	result, err := expr.Evaluate(params)
	if err != nil {
		return 0, fmt.Errorf("error evaluating expression: %w", err)
	}

	return result.(float64), nil
}

func (ee *ExpressionEvaluator) getParameters(player *Player, game *Game) map[string]interface{} {
	params := make(map[string]interface{})

	for name, amount := range player.State.Resources {
		params[name] = amount
		params[name+":max"] = player.State.ResourceMaxes[name]
		params[name+":earned"] = player.State.ResourceEarned[name]
		params[name+":ps"] = player.State.ResourcePerSecond[name]
	}

	for name, amount := range player.State.Buildings {
		params[name] = float64(amount)
	}

	for name, owned := range player.State.Upgrades {
		params[name] = boolToFloat(owned)
	}

	for name, achieved := range player.State.Achievements {
		params[name] = boolToFloat(achieved)
	}

	params["ItemsLeft"] = 100 - float64(len(player.State.Inventory))

	return params
}

func (ee *ExpressionEvaluator) have(player *Player) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("have function expects 1 argument")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("have function expects a string argument")
		}
		return boolToFloat(player.Has(key)), nil
	}
}

func (ee *ExpressionEvaluator) no(player *Player) govaluate.ExpressionFunction {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("no function expects 1 argument")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("no function expects a string argument")
		}
		return boolToFloat(!player.Has(key)), nil
	}
}

func (ee *ExpressionEvaluator) random(args ...interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("random function expects 1 or 2 arguments")
	}
	min := 0.0
	max := args[0].(float64)
	if len(args) == 2 {
		min = args[0].(float64)
		max = args[1].(float64)
	}
	return min + rand.Float64()*(max-min), nil
}

func (ee *ExpressionEvaluator) frandom(args ...interface{}) (interface{}, error) {
	return ee.random(args...)
}

func (ee *ExpressionEvaluator) chance(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("chance function expects 1 argument")
	}
	probability := args[0].(float64)
	return boolToFloat(rand.Float64()*100 < probability), nil
}

func (ee *ExpressionEvaluator) max(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("max function expects 2 arguments")
	}
	return math.Max(args[0].(float64), args[1].(float64)), nil
}

func (ee *ExpressionEvaluator) min(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("min function expects 2 arguments")
	}
	return math.Min(args[0].(float64), args[1].(float64)), nil
}

func (ee *ExpressionEvaluator) floor(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floor function expects 1 argument")
	}
	return math.Floor(args[0].(float64)), nil
}

func (ee *ExpressionEvaluator) ceil(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ceil function expects 1 argument")
	}
	return math.Ceil(args[0].(float64)), nil
}

func (ee *ExpressionEvaluator) round(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("round function expects 1 argument")
	}
	return math.Round(args[0].(float64)), nil
}

func (ee *ExpressionEvaluator) roundr(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("roundr function expects 1 argument")
	}
	value := args[0].(float64)
	fraction := value - math.Floor(value)
	if rand.Float64() < fraction {
		return math.Ceil(value), nil
	}
	return math.Floor(value), nil
}

func (ee *ExpressionEvaluator) pow(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("pow function expects 2 arguments")
	}
	return math.Pow(args[0].(float64), args[1].(float64)), nil
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (p *Player) Has(key string) bool {
	if amount, ok := p.State.Resources[key]; ok && amount > 0 {
		return true
	}
	if amount, ok := p.State.Buildings[key]; ok && amount > 0 {
		return true
	}
	if owned, ok := p.State.Upgrades[key]; ok && owned {
		return true
	}
	if achieved, ok := p.State.Achievements[key]; ok && achieved {
		return true
	}
	return false
}
