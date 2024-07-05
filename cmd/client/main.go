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
	words []string
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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.TRANSIENT,
	)
	if err != nil {
		log.Fatal(err)
	}

	gameState := gamelogic.NewGameState(username)

	for {
		input := gamelogic.GetInput()

		cmd := input[0]
		args := CommandsArgs{
			words: input,
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
