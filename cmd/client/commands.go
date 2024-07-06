package main

import (
	"fmt"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

const (
	ArmyMoves = "army_moves."
)

type command func(*gamelogic.GameState, CommandsArgs) error

func getGameCommands() map[string]command {
	return map[string]command{
		"spawn":  handlerCommandSpawn,
		"move":   handlerCommandMove,
		"status": handlerCommandStatus,
		"help":   handlerCommandHelp,
		"spam":   handlerCommandSpam,
		"quit":   handlerCommandQuit,
	}
}

func handlerCommandSpawn(gamestate *gamelogic.GameState, args CommandsArgs) error {
	return gamestate.CommandSpawn(args.words)
}

func handlerCommandMove(gamestate *gamelogic.GameState, args CommandsArgs) error {
	move, err := gamestate.CommandMove(args.words)
	if err != nil {
		return err
	}

	err = pubsub.PublishJSON(
		args.connChan,
		routing.ExchangePerilTopic,
		ArmyMoves+args.username,
		move,
	)

	return err
}

func handlerCommandStatus(gamestate *gamelogic.GameState, _ CommandsArgs) error {
	gamestate.CommandStatus()
	return nil
}

func handlerCommandHelp(_ *gamelogic.GameState, _ CommandsArgs) error {
	gamelogic.PrintClientHelp()
	return nil
}

func handlerCommandSpam(_ *gamelogic.GameState, _ CommandsArgs) error {
	fmt.Println("Spamming not allowed yet!")
	return nil
}

func handlerCommandQuit(_ *gamelogic.GameState, _ CommandsArgs) error {
	gamelogic.PrintQuit()
	os.Exit(0)
	return nil
}
