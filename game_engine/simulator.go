package game_engine

import (
	"fmt"
	"log"
	"time"
)

type GameSimulator struct {
	game *Game
}

func NewGameSimulator(game *Game) *GameSimulator {
	return &GameSimulator{
		game: game,
	}
}

func (gs *GameSimulator) SimulatePlayerProgress(player *Player, days int) map[string]uint64 {
	for i := 0; i < days; i++ {
		gs.simulateDay(player)
		log.Printf("   Buildings: %v", player.GetBuildings())
		log.Printf("   Upgrades: %+v", player.GetUpgrades())
		log.Printf("   RPS: %+v", player.State.RPS)
	}

	return player.State.Resources
}

func (gs *GameSimulator) simulateDay(player *Player) {
	log.Println("Simulating day", player.ID)
	// Симуляция покупки зданий
	buildings := gs.game.ContentSystem.GetAllContent("buildings")
	for name, _ := range buildings {
		gs.game.Buy(player, name)
	}

	// Симуляция покупки улучшений
	upgrades := gs.game.ContentSystem.GetAllContent("upgrades")
	for name, _ := range upgrades {
		gs.game.Buy(player, name)
	}

	// Обновление состояния игрока
	gs.game.updatePlayer(player)

	// Симуляция прошедшего времени
	player.State.LastSaveTime = player.State.LastSaveTime.Add(24 * time.Hour)
}

func (gs *GameSimulator) RunSimulation(numPlayers, days int) {
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("sim_player_%d", i)
		player := NewPlayer(playerID, gs.game.ContentSystem)
		gs.SimulatePlayerProgress(player, days)
		resources := player.GetResources()
		log.Printf("Simulated player %s progress after %d days: %v", playerID, days, resources)
	}

}
