package main

import (
	"fmt"
	"strings"
)

// Command представляет интерфейс для всех команд в игре
type Command interface {
	Execute(player *Player, args []string) error
	Name() string
	Description() string
}

// CommandFactory представляет фабрику для создания команд
type CommandFactory interface {
	CreateCommand(commandName string) Command
}

// GameCommandFactory реализует фабрику команд для игры
type GameCommandFactory struct {
	game *Game
}

// CreateCommand создает новую команду по имени
func (f *GameCommandFactory) CreateCommand(commandName string) Command {
	switch strings.ToLower(commandName) {
	case "buy":
		return &BuyCommand{game: f.game}
	case "sell":
		return &SellCommand{game: f.game}
	case "prestige":
		return &PrestigeCommand{game: f.game}
	case "status":
		return &StatusCommand{}
	case "help":
		return &HelpCommand{commandSystem: f.game.CommandSystem}
	case "listresources":
		return &ListResourcesCommand{}
	case "listbuildings":
		return &ListBuildingsCommand{}
	default:
		return nil
	}
}

// CommandSystem представляет систему команд в игре
type CommandSystem struct {
	commands map[string]Command
	factory  CommandFactory
}

// NewCommandSystem создает новую систему команд
func NewCommandSystem(factory CommandFactory) *CommandSystem {
	return &CommandSystem{
		commands: make(map[string]Command),
		factory:  factory,
	}
}

// RegisterCommand регистрирует новую команду в системе
func (cs *CommandSystem) RegisterCommand(cmd Command) {
	cs.commands[strings.ToLower(cmd.Name())] = cmd
}

// ExecuteCommand выполняет команду по имени
func (cs *CommandSystem) ExecuteCommand(player *Player, commandName string, args []string) error {
	cmd, exists := cs.commands[strings.ToLower(commandName)]
	if !exists {
		cmd = cs.factory.CreateCommand(commandName)
		if cmd == nil {
			return fmt.Errorf("unknown command: %s", commandName)
		}
		cs.RegisterCommand(cmd)
	}
	return cmd.Execute(player, args)
}

// GetCommandList возвращает список всех доступных команд
func (cs *CommandSystem) GetCommandList() string {
	var commandList strings.Builder
	for _, cmd := range cs.commands {
		commandList.WriteString(fmt.Sprintf("%s - %s\n", cmd.Name(), cmd.Description()))
	}
	return commandList.String()
}

// BuyCommand представляет команду для покупки зданий или улучшений
type BuyCommand struct {
	game *Game
}

func (c *BuyCommand) Execute(player *Player, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify what to buy")
	}
	itemName := strings.Join(args, " ")
	c.game.Buy(player, itemName)
	return nil
}

func (c *BuyCommand) Name() string {
	return "Buy"
}

func (c *BuyCommand) Description() string {
	return "Buy a building or upgrade"
}

// SellCommand представляет команду для продажи зданий
type SellCommand struct {
	game *Game
}

func (c *SellCommand) Execute(player *Player, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify what to sell")
	}
	itemName := strings.Join(args, " ")
	return c.game.Sell(player, itemName)
}

func (c *SellCommand) Name() string {
	return "Sell"
}

func (c *SellCommand) Description() string {
	return "Sell a building"
}

// PrestigeCommand представляет команду для выполнения престижа
type PrestigeCommand struct {
	game *Game
}

func (c *PrestigeCommand) Execute(player *Player, args []string) error {
	c.game.PerformPrestige(player)
	return nil
}

func (c *PrestigeCommand) Name() string {
	return "Prestige"
}

func (c *PrestigeCommand) Description() string {
	return "Perform a prestige reset"
}

// StatusCommand представляет команду для отображения статуса игрока
type StatusCommand struct{}

func (c *StatusCommand) Execute(player *Player, args []string) error {
	fmt.Printf("Player: %s\n", player.ID)
	fmt.Println("Resources:")
	for resource, amount := range player.State.Resources {
		fmt.Printf("%s: %.2f\n", resource, amount)
	}
	fmt.Printf("Prestige Level: %d\n", player.State.Prestige)
	return nil
}

func (c *StatusCommand) Name() string {
	return "Status"
}

func (c *StatusCommand) Description() string {
	return "Display player status"
}

// HelpCommand представляет команду для отображения списка доступных команд
type HelpCommand struct {
	commandSystem CommandSystemInterface
}

func (c *HelpCommand) Execute(player *Player, args []string) error {
	fmt.Println("Available commands:")
	fmt.Print(c.commandSystem.GetCommandList())
	return nil
}

func (c *HelpCommand) Name() string {
	return "Help"
}

func (c *HelpCommand) Description() string {
	return "Display list of available commands"
}

// ListResourcesCommand представляет команду для отображения списка ресурсов игрока
type ListResourcesCommand struct{}

func (c *ListResourcesCommand) Execute(player *Player, args []string) error {
	fmt.Println("Your resources:")
	for resource, amount := range player.State.Resources {
		fmt.Printf("%s: %.2f\n", resource, amount)
	}
	return nil
}

func (c *ListResourcesCommand) Name() string {
	return "ListResources"
}

func (c *ListResourcesCommand) Description() string {
	return "List all player resources"
}

// ListBuildingsCommand представляет команду для отображения списка зданий игрока
type ListBuildingsCommand struct{}

func (c *ListBuildingsCommand) Execute(player *Player, args []string) error {
	fmt.Println("Your buildings:")
	for name, building := range player.State.Buildings {
		fmt.Printf("%s: %d\n", name, building)
	}
	return nil
}

func (c *ListBuildingsCommand) Name() string {
	return "ListBuildings"
}

func (c *ListBuildingsCommand) Description() string {
	return "List all player buildings"
}
