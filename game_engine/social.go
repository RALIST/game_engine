package game_engine

import (
	"fmt"
)

type SocialSystem struct {
	game *Game
}

func NewSocialSystem(game *Game) *SocialSystem {
	return &SocialSystem{game: game}
}

func (ss *SocialSystem) Trade(fromPlayer, toPlayer *Player, offerResources, requestResources map[string]float64) error {
	if !fromPlayer.CanAfford(offerResources) {
		return fmt.Errorf("offering player doesn't have enough resources")
	}

	if !toPlayer.CanAfford(requestResources) {
		return fmt.Errorf("receiving player doesn't have enough resources")
	}

	for resource, amount := range offerResources {
		fromPlayer.RemoveItem(resource, amount)
		toPlayer.AddItem(resource, amount)
	}

	for resource, amount := range requestResources {
		toPlayer.RemoveItem(resource, amount)
		fromPlayer.AddItem(resource, amount)
	}

	ss.game.EventSystem.Emit("TradeConducted", map[string]interface{}{
		"FromPlayerID":     fromPlayer.ID,
		"ToPlayerID":       toPlayer.ID,
		"OfferResources":   offerResources,
		"RequestResources": requestResources,
	})

	return nil
}
