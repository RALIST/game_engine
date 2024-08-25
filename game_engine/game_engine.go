package game_engine

type GameEngine struct {
	game *Game
	ui   UIInterface
	db   DatabaseInterface
}

func (ge *GameEngine) CreatePlayer(playerID string) (*Player, error) {
	player := NewPlayer(playerID, ge.ContentSystem)
	err := ge.savePlayer(player)
	if err != nil {
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

	building, exists := ge.ContentSystem.GetBuilding(buildingName)
	if !exists {
		return fmt.Errorf("building not found: %s", buildingName)
	}

	if !player.CanAfford(building.Cost) {
		return fmt.Errorf("not enough resources to buy %s", buildingName)
	}

	player.SpendResources(building.Cost)
	player.AddBuilding(buildingName)

	return ge.savePlayer(player)
}

func (ge *GameEngine) GetPlayerResources(playerID string) (map[string]float64, error) {
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

unc (ge *GameEngine) savePlayer(player *Player) error {
	data, err := json.Marshal(player)
	if err != nil {
		return err
	}
	return ge.db.Set(context.Background(), "player:"+player.ID, string(data), 0).Err()
}

func (ge *GameEngine) loadPlayer(playerID string) (*Player, error) {
	data, err := ge.db.Get(context.Background(), "player:"+playerID).Result()
	if err != nil {
		return nil, err
	}

	var player Player
	err = json.Unmarshal([]byte(data), &player)
	if err != nil {
		return nil, err
	}

	return &player, nil
}
