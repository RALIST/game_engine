package game_engine

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/Knetic/govaluate"
)

type ExpressionEvaluator struct {
	player *Player
	game   *Game
}

func NewExpressionEvaluator(player *Player) *ExpressionEvaluator {
	return &ExpressionEvaluator{
		player: player,
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
		"and":     ee.and,
		"or":      ee.or,
	}

	prsExpr, err := ee.preprocessExpression(expression)
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(prsExpr, functions)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %w", err)
	}

	params := ee.getParameters(ee.player, ee.game)
	result, err := expr.Evaluate(params)

	if err != nil {
		return 0, fmt.Errorf("error evaluating expression: %w", err)
	}

	switch v := result.(type) {
	case float64:
		return v, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected result type: %T", v)
	}
}

func (ee *ExpressionEvaluator) getParameters(player *Player, game *Game) map[string]interface{} {
	params := make(map[string]interface{})

	for name, item := range player.State.Items {
		params[name] = item.Amount
		params[name+":max"] = player.State.ResourceMaxes[name]
		params[name+":earned"] = player.State.ResourceEarned[name]
		params[name+":ps"] = player.State.RPS[name]
	}

	params["ItemsLeft"] = 100 - float64(len(player.State.Inventory))

	return params
}

func (ee *ExpressionEvaluator) preprocessExpression(expression string) (string, error) {
	expression = strings.TrimSpace(expression)
	if !strings.HasPrefix(expression, "if ") {
		return expression, nil
	}

	// Remove the "if " prefix
	expression = expression[3:]

	// Find the end of the condition (matching parentheses)
	openParens := 0
	conditionEnd := -1
	for i, char := range expression {
		if char == '(' {
			openParens++
		} else if char == ')' {
			openParens--
			if openParens == 0 {
				conditionEnd = i
				break
			}
		}
	}

	if conditionEnd == -1 {
		return "", fmt.Errorf("invalid if statement: mismatched parentheses")
	}

	condition := strings.TrimSpace(expression[:conditionEnd+1])
	consequent := strings.TrimSpace(expression[conditionEnd+1:])

	// Replace 'and' with '&&' and 'or' with '||'
	condition = strings.ReplaceAll(condition, " and ", " && ")
	condition = strings.ReplaceAll(condition, " or ", " || ")
	if consequent == "" {
		consequent = "1"
	}

	return fmt.Sprintf("%s ? %s : 0", condition, consequent), nil
}

func (ee *ExpressionEvaluator) or(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("or function expects at least 2 arguments")
	}
	for _, arg := range args {
		val, ok := arg.(bool)
		if !ok {
			return nil, fmt.Errorf("or function expects boolean arguments")
		}
		if val {
			return true, nil
		}
	}
	return false, nil
}

func (ee *ExpressionEvaluator) and(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("and function expects at least 2 arguments")
	}
	for _, arg := range args {
		val, ok := arg.(bool)
		if !ok {
			return nil, fmt.Errorf("and function expects boolean arguments")
		}
		if !val {
			return false, nil
		}
	}
	return true, nil
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
	minN := 0.0
	maxN := args[0].(float64)
	if len(args) == 2 {
		minN = args[0].(float64)
		maxN = args[1].(float64)
	}

	return float64(int(minN) + rand.Intn(int(maxN-minN+1))), nil
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
