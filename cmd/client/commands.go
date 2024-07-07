package main

import (
	"errors"
	"os"
	"strconv"
	"time"

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

func handlerCommandSpam(_ *gamelogic.GameState, args CommandsArgs) error {
	if len(args.words) < 2 {
		return errors.New("input must have at least 2 words")
	}

	num, err := strconv.Atoi(args.words[1])
	if err != nil {
		return errors.New("must be a valid number")
	}

	for range num {
		msg := gamelogic.GetMaliciousLog()
		pubsub.PublishGob(
			args.connChan,
			routing.ExchangePerilTopic,
			routing.GameLogSlug+"."+args.username,
			routing.GameLog{
				CurrentTime: time.Now().UTC(),
				Message:     msg,
				Username:    args.username,
			},
		)
	}

	return nil
}

func handlerCommandQuit(_ *gamelogic.GameState, _ CommandsArgs) error {
	gamelogic.PrintQuit()
	os.Exit(0)
	return nil
}
