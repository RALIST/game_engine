package game_engine

import (
	"time"
)

type PlayerItem struct {
	ID             string
	Type           string
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Cost           map[string]float64     `json:"cost"`
	RPS            map[string]float64     `json:"resourcePerSecond"`
	ResourceEarned map[string]uint64      `json:"resourceEarned"`
	Amount         int                    `json:"amount"`
	Effects        []Effect               `json:"effects"`
	Reqs           []string               `json:"reqs"`
	Properties     map[string]interface{} `json:"properties"`
}

// PlayerState представляет текущее состояние игрока
type PlayerState struct {
	Resources         map[string]uint64      `json:"resources"`
	Upgrades          map[string]bool        `json:"upgrades"`
	Buildings         map[string]int         `json:"buildings"`
	Achievements      map[string]bool        `json:"achievements"`
	Shinies           map[string]ShinyState  `json:"shinies"`
	Prestige          int                    `json:"prestige"`
	LastSaveTime      time.Time              `json:"lastSaveTime"`
	Log               []string               `json:"log"`
	AchievementLevels map[string]int         `json:"achievementLevels"`
	Items             map[string]*PlayerItem `json:"data"`
	ResourceMaxes     map[string]uint64      `json:"resourceMaxes"`
	ResourceEarned    map[string]uint64      `json:"resourceEarned"`
	RPS               map[string]int         `json:"resourcePerSecond"`
	Inventory         []string               `json:"inventory"`
}

// ShinyState представляет состояние "блестящего" объекта
type ShinyState struct {
	Active    bool      `json:"active"`
	LastSpawn time.Time `json:"lastSpawn"`
}

// Player представляет игрока в игре
type Player struct {
	ID                string       `json:"id"`
	State             *PlayerState `json:"state"`
	Config            *ContentSystem
	ResourcePerSecond interface{}
}

// NewPlayer создает нового игрока с заданным ID и конфигурацией
func NewPlayer(playerID string, cfg *ContentSystem) *Player {
	p := &Player{
		ID: playerID,
		State: &PlayerState{
			Achievements:      make(map[string]bool),
			Shinies:           make(map[string]ShinyState),
			Prestige:          0,
			Items:             initItems(cfg),
			LastSaveTime:      time.Now(),
			AchievementLevels: make(map[string]int),
			Log:               make([]string, 0, 10),
		},
		Config: cfg,
	}

	p.RecalculateState()
	return p
}

func initItems(cfg *ContentSystem) map[string]*PlayerItem {
	items := make(map[string]*PlayerItem)
	for _, item := range cfg.Items {
		playerItem := &PlayerItem{
			ID:          item.ID,
			Type:        item.Type,
			Name:        item.Name,
			Description: item.Description,
			Cost:        item.Cost,
			Reqs:        item.Reqs,
			Properties:  item.Properties,
			Amount:      item.Initial,
			Effects:     item.Effects,
		}
		items[item.ID] = playerItem
	}

	return items
}

func (p *Player) RecalculateState() {
	rps := map[string]int{}
	for _, item := range p.State.Items {
		if item.Amount == 0 {
			continue
		}

		for _, effect := range item.Effects {
			if effect.Type == "yield" {
				value, _ := evaluateExpression(p, effect.Expression)
				rps[effect.Target] += int(value)
			}
		}
	}

	p.State.RPS = rps
}

// AddItem добавляет ресурсы игроку
func (p *Player) AddItem(itemID string, amount float64) {
	item := p.State.Items[itemID]
	if item.Type == "buildings" {
		item.Amount += int(amount)
	} else {
		if item.Amount > 0 {
			return
		}

		item.Amount += int(amount)
	}
}

// RemoveItem удаляет ресурсы у игрока
func (p *Player) RemoveItem(itemID string, amount float64) {
	p.State.Items[itemID].Amount -= int(amount)
}

// CanAfford проверяет, может ли игрок позволить себе покупку
func (p *Player) CanAfford(cost map[string]float64) bool {
	for resource, amount := range cost {
		res := p.State.Items[resource]
		if res.Amount < int(amount) || amount == 0 {
			return false
		}
	}

	return true
}

// SpendResources тратит ресурсы игрока
func (p *Player) SpendResources(cost map[string]float64) {
	for resource, amount := range cost {
		p.State.Items[resource].Amount -= int(amount)
	}
}

func (p *Player) GetResources() map[string]float64 {
	resources := make(map[string]float64)
	for id, item := range p.State.Items {
		if item.Type == "resources" {
			resources[id] = float64(item.Amount)
		}
	}

	return resources
}

func (p *Player) GetBuildings() map[string]float64 {
	resources := make(map[string]float64)
	for id, item := range p.State.Items {
		if item.Type == "buildings" {
			resources[id] = float64(item.Amount)
		}
	}

	return resources
}

func (p *Player) GetUpgrades() map[string]float64 {
	resources := make(map[string]float64)
	for id, item := range p.State.Items {
		if item.Type == "upgrades" {
			resources[id] = float64(item.Amount)
		}
	}

	return resources
}

// AddLog добавляет сообщение в лог игрока
func (p *Player) AddLog(message string) {
	p.State.Log = append(p.State.Log, message)
	if len(p.State.Log) > 10 {
		p.State.Log = p.State.Log[1:]
	}
}

// ResetProgress сбрасывает прогресс игрока и увеличивает уровень престижа
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

// GetItemAmount возвращает количество зданий определенного типа
func (p *Player) GetItemAmount(itemID string) int {
	for _, item := range p.State.Items {
		if item.ID == itemID {
			return item.Amount
		}
	}

	return 0
}

func (p *Player) GetItem(itemID string) *PlayerItem {
	return p.State.Items[itemID]
}

// HasUpgrade проверяет, есть ли у игрока определенное улучшение
func (p *Player) HasUpgrade(upgradeName string) bool {
	return p.State.Upgrades[upgradeName]
}

// GetAchievementStatus проверяет, получено ли определенное достижение
func (p *Player) GetAchievementStatus(achievementName string) bool {
	return p.State.Achievements[achievementName]
}

// SetAchievement устанавливает статус достижения
func (p *Player) SetAchievement(achievementName string) {
	p.State.Achievements[achievementName] = true
}

// GetShinyState возвращает состояние "блестящего" объекта
func (p *Player) GetShinyState(shinyName string) ShinyState {
	return p.State.Shinies[shinyName]
}

// SetShinyState устанавливает состояние "блестящего" объекта
func (p *Player) SetShinyState(shinyName string, state ShinyState) {
	p.State.Shinies[shinyName] = state
}

// GetAchievementLevel возвращает уровень достижения
func (p *Player) GetAchievementLevel(achievementName string) int {
	return p.State.AchievementLevels[achievementName]
}

// SetAchievementLevel устанавливает уровень достижения
func (p *Player) SetAchievementLevel(achievementName string, level int) {
	p.State.AchievementLevels[achievementName] = level
}
