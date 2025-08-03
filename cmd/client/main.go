package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const CONN_STRING = "amqp://guest:guest@localhost:5672/"

func main() {
	conn, err := amqp.Dial(CONN_STRING)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Starting Peril client...")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Printf("could not get username: %v", err)
		log.Fatal(err)
	}

	pubsub.DeclareAndBind(conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
	)

	gameState := gamelogic.NewGameState(username)

OuterLoop:
	for {
		wordSlice := gamelogic.GetInput()
		if len(wordSlice) == 0 {
			continue
		}

		switch wordSlice[0] {
		case "spawn":
			err := gameState.CommandSpawn(wordSlice)
			if err != nil {
				log.Printf("error spawning: %v\n", err)
				log.Fatal(err)
			}
		case "move":
			_, err := gameState.CommandMove(wordSlice)
			if err != nil {
				fmt.Printf("error moving: %v\n", err)
				log.Fatal(err)
			}
			fmt.Println("Sucess moving")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			fmt.Println("Quitting...")
			break OuterLoop
		default:
			fmt.Println("Unknown command")
		}

	}

	// wait for ctrl+c
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	// fmt.Println("Shutting down gracefully...")

}
