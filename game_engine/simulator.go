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
		log.Println("Simulating day", i)
		gs.simulateDay(player)
		log.Printf("Simulated player %s progress at %d day: %v", player.ID, i, player.State.Resources)
	}

	return player.State.Resources
}

func (gs *GameSimulator) simulateDay(player *Player) {
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

	// Симуляция появления редких событий
	//shinies := gs.game.ContentSystem.GetAllContent("shinies")
	//for name, shiny := range shinies {
	//	frequency, ok := shiny.Properties["frequency"].(float64)
	//	if !ok {
	//		// Если frequency не float64, пытаемся преобразовать из int
	//		if freqInt, ok := shiny.Properties["frequency"].(int); ok {
	//			frequency = float64(freqInt)
	//		} else {
	//			log.Printf("Warning: Invalid frequency for shiny %s", name)
	//			continue
	//		}
	//	}
	//	if rand.Float64() < frequency {
	//		gs.game.applyShinyEffect(player, name)
	//	}
	//}

	// Обновление состояния игрока
	gs.game.updatePlayer(player)

	// Симуляция прошедшего времени
	player.State.LastSaveTime = player.State.LastSaveTime.Add(24 * time.Hour)
}

func (gs *GameSimulator) RunSimulation(numPlayers, days int) {
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("sim_player_%d", i)
		player := NewPlayer(playerID, gs.game.ContentSystem)
		resources := gs.SimulatePlayerProgress(player, days)
		log.Printf("Simulated player %s progress after %d days: %v", playerID, days, resources)
	}

}
