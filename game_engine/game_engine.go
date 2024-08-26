package game_engine

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ralist/game_engine/game_engine/config"
)

// GameEngine is a game engine responsible for operations with databases and players
type GameEngine struct {
	Game *Game
	db   DatabaseInterface
}

func NewGameEngine(fileName string, db DatabaseInterface) (*GameEngine, error) {
	// Load game configuration
	cfg, err := config.LoadConfig(fileName)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	game, err := NewGame(cfg)
	engine := &GameEngine{
		Game: game,
		db:   db,
	}
	return engine, err
}

func (ge *GameEngine) Run() {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if err := ge.updatePlayers(); err != nil {
				log.Printf("Error updating players: %v", err)
			}
		}
	}
}

func (ge *GameEngine) updatePlayers() error {
	log.Println("Updating players")
	players, err := ge.db.LoadPlayers()
	if err != nil {
		return fmt.Errorf("error loading players: %w", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(players))

	for _, playerData := range players {
		wg.Add(1)
		go func(data interface{}) {
			defer wg.Done()
			var player Player
			if err := json.Unmarshal(data.([]byte), &player); err != nil {
				errCh <- fmt.Errorf("error unmarshaling player data: %w", err)
				return
			}
			ge.updatePlayer(&player)
			if err := ge.savePlayer(&player); err != nil {
				errCh <- fmt.Errorf("error saving player: %w", err)
			}
		}(playerData)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			log.Printf("Error during player update: %v", err)
		}
	}

	return nil
}

func (ge *GameEngine) updatePlayer(player *Player) {
	ge.Game.updatePlayer(player)
}

func (ge *GameEngine) CreatePlayer(playerID string) (*Player, error) {
	player := NewPlayer(playerID, ge.Game.ContentSystem)
	if err := ge.savePlayer(player); err != nil {
		return nil, fmt.Errorf("error creating player: %w", err)
	}
	return player, nil
}

func (ge *GameEngine) GetPlayer(playerID string) (*Player, error) {
	return ge.loadPlayer(playerID)
}

func (ge *GameEngine) BuyBuilding(playerID, buildingName string) error {
	player, err := ge.loadPlayer(playerID)
	if err != nil {
		return fmt.Errorf("error loading player: %w", err)
	}

	ge.Game.Buy(player, buildingName)
	if err := ge.savePlayer(player); err != nil {
		return fmt.Errorf("error saving player after buying building: %w", err)
	}
	return nil
}

func (ge *GameEngine) GetPlayerResources(playerID string) (map[string]uint64, error) {
	player, err := ge.loadPlayer(playerID)
	if err != nil {
		return nil, fmt.Errorf("error loading player resources: %w", err)
	}
	return player.State.Resources, nil
}

func (ge *GameEngine) GetPlayerBuildings(playerID string) (map[string]int, error) {
	player, err := ge.loadPlayer(playerID)
	if err != nil {
		return nil, fmt.Errorf("error loading player buildings: %w", err)
	}
	return player.State.Buildings, nil
}

func (ge *GameEngine) savePlayer(player *Player) error {
	data, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("error marshaling player data: %w", err)
	}
	if err := ge.db.SavePlayer(player.ID, data); err != nil {
		return fmt.Errorf("error saving player to database: %w", err)
	}
	return nil
}

func (ge *GameEngine) loadPlayer(playerID string) (*Player, error) {
	data, err := ge.db.LoadPlayer(playerID)
	if err != nil {
		return nil, fmt.Errorf("error loading player from database: %w", err)
	}

	var player Player
	if err := json.Unmarshal(data, &player); err != nil {
		return nil, fmt.Errorf("error unmarshaling player data: %w", err)
	}

	return &player, nil
}
