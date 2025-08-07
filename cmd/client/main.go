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
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

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
		pubsub.SimpleQueueTransient,
	)

	gameState := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, channel),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

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
			move, err := gameState.CommandMove(wordSlice)
			if err != nil {
				fmt.Printf("error moving: %v\n", err)
				continue
			}

			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
				move,
			)
			if err != nil {
				fmt.Printf("error publishing message: %v", err)
				continue
			}
			fmt.Println("Success moving")
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
}
