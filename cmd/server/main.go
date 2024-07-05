package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	connstr := os.Getenv("CONN_STR")

	conn, err := amqp.Dial(connstr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Sucessfully connected")

	connChan, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	pubsub.PublishJSON(connChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
		IsPaused: true,
	})

	fmt.Println("Shutting down program")
}
