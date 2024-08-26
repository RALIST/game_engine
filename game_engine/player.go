package game_engine

import (
	"time"
)

const maxLogEntries = 10

type PlayerState struct {
	Resources         map[string]float64    `json:"resources"`
	Upgrades          map[string]bool       `json:"upgrades"`
	Buildings         map[string]float64    `json:"buildings"`
	Achievements      map[string]bool       `json:"achievements"`
	Shinies           map[string]ShinyState `json:"shinies"`
	Prestige          int                   `json:"prestige"`
	LastSaveTime      time.Time             `json:"lastSaveTime"`
	Log               []string              `json:"log"`
	AchievementLevels map[string]int        `json:"achievementLevels"`
}

// ShinyState представляет текущее состояние редкого события
type ShinyState struct {
	Active    bool      `json:"active"`
	LastSpawn time.Time `json:"lastSpawn"`
}

// Player представляет игрока
type Player struct {
	ID     string      `json:"id"`
	State  PlayerState `json:"state"`
	Config *ContentSystem
}

// NewPlayer создает нового игрока
func NewPlayer(playerID string, cfg *ContentSystem) *Player {
	return &Player{
		ID: playerID,
		State: PlayerState{
			Resources:         initResources(cfg),
			Upgrades:          make(map[string]bool),
			Buildings:         initBuildings(cfg),
			Achievements:      make(map[string]bool),
			Shinies:           make(map[string]ShinyState),
			Prestige:          0,
			LastSaveTime:      time.Now(),
			AchievementLevels: make(map[string]int),
			Log:               make([]string, 0),
		},
		Config: cfg,
	}
}

func initResources(cfg *ContentSystem) map[string]float64 {
	resources := make(map[string]float64)
	for name, resource := range cfg.Resources {
		resources[name] = resource.Initial
	}
	return resources
}

func initBuildings(cfg *ContentSystem) map[string]float64 {
	buildings := make(map[string]float64)
	for name, building := range cfg.Buildings {
		buildings[name] = building.Initial
	}
	return buildings
}

// AddUpgrade добавляет улучшение игроку
func (p *Player) AddUpgrade(upgradeName string) {
	p.State.Upgrades[upgradeName] = true
}

// AddBuilding добавляет здание игроку
func (p *Player) AddBuilding(buildingName string) {
	p.State.Buildings[buildingName]++
}

// AddResource добавляет ресурс игроку
func (p *Player) AddResource(resourceName string, amount float64) {
	p.State.Resources[resourceName] += amount
}

// RemoveResource удаляет ресурс у игрока
func (p *Player) RemoveResource(resourceName string, amount float64) {
	p.State.Resources[resourceName] -= amount
	if p.State.Resources[resourceName] < 0 {
		p.State.Resources[resourceName] = 0
	}
}

// CanAfford проверяет, может ли игрок позволить себе покупку
func (p *Player) CanAfford(cost map[string]float64) bool {
	for resource, amount := range cost {
		if p.State.Resources[resource] < amount {
			return false
		}
	}
	return true
}

// SpendResources тратит ресурсы игрока на покупку
func (p *Player) SpendResources(cost map[string]float64) {
	for resource, amount := range cost {
		p.RemoveResource(resource, amount)
	}
}

// AddLog добавляет сообщение в лог игрока
func (p *Player) AddLog(message string) {
	p.State.Log = append(p.State.Log, message)
	if len(p.State.Log) > 10 {
		p.State.Log = p.State.Log[len(p.State.Log)-10:]
	}
}

// ResetProgress сбрасывает прогресс игрока при престиже
func (p *Player) ResetProgress() {
	p.State.Prestige++
	for resource := range p.State.Resources {
		p.State.Resources[resource] = p.Config.Resources[resource].Initial
	}
	for name := range p.State.Buildings {
		p.State.Buildings[name] = 0
	}
	p.AddLog("You have prestiged! All your progress has been reset, but you now earn more resources.")
}

// GetBuildingAmount возвращает количество зданий определенного типа
func (p *Player) GetBuildingAmount(buildingName string) float64 {
	return p.State.Buildings[buildingName]
}

// HasUpgrade проверяет, есть ли у игрока определенное улучшение
func (p *Player) HasUpgrade(upgradeName string) bool {
	return p.State.Upgrades[upgradeName]
}

// GetAchievementStatus возвращает статус достижения
func (p *Player) GetAchievementStatus(achievementName string) bool {
	return p.State.Achievements[achievementName]
}

// SetAchievement устанавливает статус достижения
func (p *Player) SetAchievement(achievementName string) {
	p.State.Achievements[achievementName] = true
}

// GetShinyState возвращает состояние редкого события
func (p *Player) GetShinyState(shinyName string) ShinyState {
	return p.State.Shinies[shinyName]
}

// SetShinyState устанавливает состояние редкого события
func (p *Player) SetShinyState(shinyName string, state ShinyState) {
	p.State.Shinies[shinyName] = state
}

func (p *Player) GetAchievementLevel(achievementName string) int {
	// Предполагаем, что уровни достижений хранятся в map[string]int
	if level, ok := p.State.AchievementLevels[achievementName]; ok {
		return level
	}
	return 0
}

func (p *Player) SetAchievementLevel(achievementName string, level int) {
	p.State.AchievementLevels[achievementName] = level
}
