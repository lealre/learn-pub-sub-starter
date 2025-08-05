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

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	pubsub.DeclareAndBind(conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
	)

	gameState := gamelogic.NewGameState(username)
	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.Transient, handlerPause(gameState))

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
				continue
			}
		case "move":
			_, err := gameState.CommandMove(wordSlice)
			if err != nil {
				fmt.Printf("error moving: %v\n", err)
				continue
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
