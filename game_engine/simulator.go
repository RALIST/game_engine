package game_engine

import (
	"math/rand"
)

type GameSimulator struct {
	game *Game
}

func NewGameSimulator(game *Game) *GameSimulator {
	return &GameSimulator{game: game}
}

func (gs *GameSimulator) SimulatePlayerProgress(days int) map[string]float64 {
	player := NewPlayer("sim_player", gs.game.State.Config)

	for i := 0; i < days; i++ {
		gs.simulateDay(player)
	}

	return player.State.Resources
}

func (gs *GameSimulator) simulateDay(player *Player) {
	// Симуляция действий игрока: покупка зданий, улучшений и т.д.
	// Это упрощенная версия, вам нужно будет адаптировать ее под вашу игровую логику
	for buildingName := range gs.game.State.Config.Buildings {
		if rand.Float64() < 0.1 { // 10% шанс купить здание
			gs.game.Buy(player, buildingName)
		}
	}

	for upgradeName := range gs.game.State.Config.Upgrades {
		if rand.Float64() < 0.05 { // 5% шанс купить улучшение
			gs.game.Buy(player, upgradeName)
		}
	}

	gs.game.State.Players[player.ID] = player
}
