package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/Knetic/govaluate"
)

type ExpressionEvaluator struct {
	variables map[string]float64
}

func NewExpressionEvaluator() *ExpressionEvaluator {
	return &ExpressionEvaluator{
		variables: make(map[string]float64),
	}
}

func (ee *ExpressionEvaluator) SetVariable(name string, value float64) {
	ee.variables[name] = value
}

func (ee *ExpressionEvaluator) Evaluate(expression string) (float64, error) {
	exp, err := govaluate.NewEvaluableExpression(expression)
	params := make(map[string]interface{})
	for k, v := range ee.variables {
		params[k] = v
	}
	result, err := exp.Evaluate(params)
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}
func (ee *ExpressionEvaluator) eval(expr ast.Expr) (float64, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		left, err := ee.eval(e.X)
		if err != nil {
			return 0, err
		}
		right, err := ee.eval(e.Y)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
	case *ast.BasicLit:
		return strconv.ParseFloat(e.Value, 64)
	case *ast.Ident:
		if value, ok := ee.variables[e.Name]; ok {
			return value, nil
		}
		return 0, fmt.Errorf("undefined variable: %s", e.Name)
	}
	return 0, fmt.Errorf("unsupported expression type")
}
