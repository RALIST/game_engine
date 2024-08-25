package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/yourusername/idle-game-engine/config"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	gameConfig, err := config.LoadConfig("config/gold_rush_config.yaml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Создаем JSON-базу данных
	db := NewJSONDatabase("game_data.json")

	// Создаем игровой движок
	gameEngine := NewGameEngine(gameConfig, &DummyUI{}, db)

	// Создаем фабрику команд
	commandFactory := &GameCommandFactory{game: gameEngine.game}

	// Инициализируем систему команд с использованием фабрики
	commandSystem := NewCommandSystem(commandFactory)

	// Устанавливаем систему команд в игру
	gameEngine.game.CommandSystem = commandSystem

	// Добавляем обработчики событий
	gameEngine.game.EventSystem.On("BuildingBought", func(data map[string]interface{}) {
		fmt.Printf("Player %s bought building %s. New amount: %d\n",
			data["PlayerID"], data["BuildingName"], data["Amount"])
	})

	gameEngine.game.EventSystem.On("UpgradeBought", func(data map[string]interface{}) {
		fmt.Printf("Player %s bought upgrade %s\n",
			data["PlayerID"], data["UpgradeName"])
	})

	gameEngine.game.EventSystem.On("BuildingSold", func(data map[string]interface{}) {
		fmt.Printf("Player %s sold building %s. New amount: %d\n",
			data["PlayerID"], data["BuildingName"], data["Amount"])
	})

	gameEngine.game.EventSystem.On("Prestige", func(data map[string]interface{}) {
		fmt.Printf("Player %s performed prestige. New level: %d\n",
			data["PlayerID"], data["PrestigeLevel"])
	})

	gameEngine.game.EventSystem.On("AchievementUnlocked", func(data map[string]interface{}) {
		fmt.Printf("Player %s unlocked achievement: %s\n",
			data["PlayerID"], data["AchievementName"])
	})

	// Настраиваем и запускаем HTTP-сервер
	r := setupServer(gameEngine)
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
