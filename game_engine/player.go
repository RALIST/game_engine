package game_engine

import (
	"time"
)

type PlayerState struct {
	Resources         map[string]uint64      `json:"resources"`
	Multipliers       map[string]float64     `json:"multipliers"`
	Upgrades          map[string]bool        `json:"upgrades"`
	Buildings         map[string]int         `json:"buildings"`
	Achievements      map[string]bool        `json:"achievements"`
	Shinies           map[string]ShinyState  `json:"shinies"`
	Prestige          int                    `json:"prestige"`
	LastSaveTime      time.Time              `json:"lastSaveTime"`
	Log               []string               `json:"log"`
	AchievementLevels map[string]int         `json:"achievementLevels"`
	Data              map[string]ContentItem `json:"data"`
}

type ShinyState struct {
	Active    bool      `json:"active"`
	LastSpawn time.Time `json:"lastSpawn"`
}

type Player struct {
	ID     string      `json:"id"`
	State  PlayerState `json:"state"`
	Config *ContentSystem
}

func NewPlayer(playerID string, cfg *ContentSystem) *Player {
	return &Player{
		ID: playerID,
		State: PlayerState{
			Resources:         initResources(cfg),
			Multipliers:       initMultipliers(cfg),
			Upgrades:          initUpgrades(cfg),
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

func initResources(cfg *ContentSystem) map[string]uint64 {
	resources := make(map[string]uint64)
	for name, resource := range cfg.GetAllContent("resources") {
		resources[name] = uint64(resource.Initial)
	}
	return resources
}

func initUpgrades(cfg *ContentSystem) map[string]bool {
	upgrades := make(map[string]bool)
	for name, _ := range cfg.GetAllContent("upgrades") {
		upgrades[name] = false
	}
	return upgrades
}

func initMultipliers(cfg *ContentSystem) map[string]float64 {
	multipliers := make(map[string]float64)
	for name, _ := range cfg.GetAllContent("resources") {
		multipliers[name] = 1
	}

	return multipliers
}

func initBuildings(cfg *ContentSystem) map[string]int {
	buildings := make(map[string]int)
	for name, building := range cfg.GetAllContent("buildings") {
		buildings[name] = building.Initial
	}
	return buildings
}

func (p *Player) AddUpgrade(upgradeName string) {
	p.State.Upgrades[upgradeName] = true
}

func (p *Player) AddBuilding(buildingName string) {
	p.State.Buildings[buildingName]++
}

func (p *Player) AddResource(resourceName string, amount float64) {
	p.State.Resources[resourceName] += uint64(amount)
}

func (p *Player) RemoveResource(resourceName string, amount float64) {
	p.State.Resources[resourceName] -= uint64(amount)
	if p.State.Resources[resourceName] < 0 {
		p.State.Resources[resourceName] = 0
	}
}

func (p *Player) CanAfford(cost map[string]float64) bool {
	for resource, amount := range cost {
		if p.State.Resources[resource] < uint64(amount) {
			return false
		}
	}
	return true
}

func (p *Player) SpendResources(cost map[string]float64) {
	for resource, amount := range cost {
		p.RemoveResource(resource, amount)
	}
}

func (p *Player) AddLog(message string) {
	p.State.Log = append(p.State.Log, message)
	if len(p.State.Log) > 10 {
		p.State.Log = p.State.Log[len(p.State.Log)-10:]
	}
}

func (p *Player) ResetProgress() {
	p.State.Prestige++
	resources := p.Config.GetAllContent("resources")
	for resource := range p.State.Resources {
		if resourceItem, ok := resources[resource]; ok {
			p.State.Resources[resource] = uint64(resourceItem.Initial)
		} else {
			p.State.Resources[resource] = 0
		}
	}
	for name := range p.State.Buildings {
		p.State.Buildings[name] = 0
	}
	p.AddLog("You have prestiged! All your progress has been reset, but you now earn more resources.")
}

func (p *Player) GetBuildingAmount(buildingName string) int {
	return p.State.Buildings[buildingName]
}

func (p *Player) HasUpgrade(upgradeName string) bool {
	return p.State.Upgrades[upgradeName]
}

func (p *Player) GetAchievementStatus(achievementName string) bool {
	return p.State.Achievements[achievementName]
}

func (p *Player) SetAchievement(achievementName string) {
	p.State.Achievements[achievementName] = true
}

func (p *Player) GetShinyState(shinyName string) ShinyState {
	return p.State.Shinies[shinyName]
}

func (p *Player) SetShinyState(shinyName string, state ShinyState) {
	p.State.Shinies[shinyName] = state
}

func (p *Player) GetAchievementLevel(achievementName string) int {
	if level, ok := p.State.AchievementLevels[achievementName]; ok {
		return level
	}
	return 0
}

func (p *Player) SetAchievementLevel(achievementName string, level int) {
	p.State.AchievementLevels[achievementName] = level
}

func (p *Player) getVariables() map[string]float64 {
	variables := make(map[string]float64)
	for resource, amount := range p.State.Resources {
		variables[resource] = float64(amount)
	}
	for building, amount := range p.State.Buildings {
		variables[building] = float64(amount)
	}
	variables["prestige"] = float64(p.State.Prestige)
	return variables
}
