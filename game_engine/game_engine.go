package game_engine

import (
	"encoding/json"
	"github.com/ralist/game_engine/game_engine/config"
	"log"
	"time"
)

type GameEngine struct {
	Game *Game
	ui   UIInterface
	db   DatabaseInterface
}

func NewGameEngine(fileName string, db DatabaseInterface) *GameEngine {
	// Load game configuration
	cfg, err := config.LoadConfig(fileName)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	game := NewGame(cfg)
	engine := &GameEngine{
		Game: game,
		db:   db,
	}
	return engine
}

func (ge *GameEngine) Run() {
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-t.C:
			ge.updatePlayers()
		}
	}
}

func (ge *GameEngine) updatePlayers() {
	log.Println("Updating players")
	players, err := ge.db.LoadPlayers()
	if err != nil {
		log.Fatalf("Error loading players: %v", err)
	}

	for _, playerData := range players {
		go func() {
			var player Player
			err = json.Unmarshal(playerData.([]byte), &player)
			ge.updatePlayer(&player)
			err := ge.savePlayer(&player)
			if err != nil {
				return
			}
		}()
	}
}

func (ge *GameEngine) updatePlayer(player *Player) {
	ge.Game.updatePlayer(player)
}

func (ge *GameEngine) CreatePlayer(playerID string) (*Player, error) {
	player := NewPlayer(playerID, ge.Game.ContentSystem)
	err := ge.savePlayer(player)
	if err != nil {
		log.Println("Error creating player:", err)
		return nil, err
	}
	return player, nil
}

func (ge *GameEngine) GetPlayer(playerID string) (*Player, error) {
	return ge.loadPlayer(playerID)
}

func (ge *GameEngine) BuyBuilding(playerID, buildingName string) error {
	player, err := ge.loadPlayer(playerID)
	if err != nil {
		return err
	}

	ge.Game.Buy(player, buildingName)
	return ge.savePlayer(player)
}

func (ge *GameEngine) GetPlayerResources(playerID string) (map[string]uint64, error) {
	player, err := ge.loadPlayer(playerID)
	if err != nil {
		return nil, err
	}
	return player.State.Resources, nil
}

func (ge *GameEngine) GetPlayerBuildings(playerID string) (map[string]int, error) {
	player, err := ge.loadPlayer(playerID)
	if err != nil {
		return nil, err
	}
	return player.State.Buildings, nil
}

func (ge *GameEngine) savePlayer(player *Player) error {
	data, err := json.Marshal(player)
	err = ge.db.SavePlayer(player.ID, data)
	if err != nil {
		return err
	}

	return nil
}

func (ge *GameEngine) loadPlayer(playerID string) (*Player, error) {
	data, err := ge.db.LoadPlayer(playerID)
	if err != nil {
		return nil, err
	}

	var player Player
	err = json.Unmarshal(data, &player)
	if err != nil {
		return nil, err
	}

	return &player, nil
}
