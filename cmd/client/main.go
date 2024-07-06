package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CommandsArgs struct {
	words    []string
	username string
	connChan *amqp.Channel
}

type HandlerArgs struct {
	connChan *amqp.Channel
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting Peril client...")

	connstr := os.Getenv("CONN_STR")

	conn, err := amqp.Dial(connstr)
	if err != nil {
		log.Fatal(err)
	}

	connChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(username)

	handlerArgs := HandlerArgs{
		connChan: connChan,
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.TRANSIENT,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.TRANSIENT,
		handlerMove(gameState, handlerArgs),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.DURABLE,
		handlerWar(gameState, handlerArgs),
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		input := gamelogic.GetInput()

		cmd := input[0]
		args := CommandsArgs{
			words:    input,
			username: gameState.GetUsername(),
			connChan: connChan,
		}

		command, exists := getGameCommands()[cmd]
		if !exists {
			fmt.Println("Command does not exist")
			continue
		}

		err := command(gameState, args)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
