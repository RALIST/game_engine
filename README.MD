# Idle Game Engine

This is a flexible and extensible idle game engine written in Go. It provides a solid foundation for creating idle/incremental games with features like resources, buildings, upgrades, achievements, and more.

## Features

- Resource management
- Building and upgrade systems
- Achievement system
- Prestige mechanic
- Expression evaluation for dynamic game mechanics
- Event system
- Command system for player interactions
- Caching system for performance optimization
- YAML-based game configuration
- Localization support
- Simulation capabilities for game balancing

## Getting Started

### Prerequisites

- Go 1.16 or higher
- Dependencies (install using `go get`):
  - `gopkg.in/yaml.v2`
  - `github.com/Knetic/govaluate`

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/ralist/game-engine.git
   ```

2. Navigate to the project directory:
   ```
   cd idle-game-engine
   ```

3. Install dependencies:
   ```
   go get -u gopkg.in/yaml.v2
   go get -u github.com/Knetic/govaluate
   ```

### Configuration

1. Edit the `game_engine/config/gold_rush_config.yaml` file to define your game's resources, buildings, upgrades, achievements, and other game elements.

2. Customize the game logic in `game_engine/game.go` if needed.

### Running the Game

To run the game, create a main file (e.g., `main.go`) with the following content:

```go
package main

import (
    "log"
    "github.com/ralist/game-engine/config"
    "github.com/ralist/game-engine/game_engine"
)

func main() {
    cfg, err := config.LoadConfig("game_engine/config/gold_rush_config.yaml")
    if err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    ui := &YourUIImplementation{}
    db := &YourDatabaseImplementation{}

    engine := game_engine.NewGameEngine(cfg, ui, db)
    engine.Run()
}
```

Replace `YourUIImplementation` and `YourDatabaseImplementation` with your own implementations of the `UIInterface` and `DatabaseInterface` respectively.

## Usage

### Adding New Game Elements

1. To add new resources, buildings, upgrades, or achievements, edit the `gold_rush_config.yaml` file.

2. To add new game mechanics, extend the `Game` struct in `game.go` and add new methods as needed.

### Creating Commands

To add new player commands:

1. Create a new command struct implementing the `Command` interface in `commands.go`.
2. Add the new command to the `GameCommandFactory` in `commands.go`.

### Handling Events

To listen for game events:

1. Use the `EventSystem.On` method to register event listeners.
2. Emit events using the `EventSystem.Emit` method when appropriate.

### Localization

1. Create JSON files for each language in a `localization` directory.
2. Use the `LocalizationSystem` to load and retrieve translations.

### Simulation

Use the `GameSimulator` in `simulator.go` to run simulations for game balancing and testing.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
```

This README provides an overview of your game engine, its features, how to set it up, and how to use its various components. You may want to expand on certain sections or add more specific examples depending on how you expect people to use your engine.

Remember to replace "yourusername" with your actual GitHub username or the appropriate repository path. Also, if you haven't already, consider adding proper documentation comments to your code, especially for exported functions and types, to make it easier for others (and yourself) to understand and use the engine.
