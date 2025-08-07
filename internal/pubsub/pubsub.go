package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const (
	Durable   simpleQueueType = "duarble"
	Transient simpleQueueType = "transient"
)

type Acktype string

const (
	Ack         Acktype = "Ack"
	NackRequeue Acktype = "NackRequeue"
	NackDiscard Acktype = "NackDiscard"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes,
		},
	)

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	amqoTable := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	queue, err := channel.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		amqoTable,
	)

	if err != nil {
		log.Printf("error declaring queue: %v", err)
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)

	if err != nil {
		log.Printf("error binding queue: %v", err)
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType,
	handler func(T) Acktype,
) error {

	chann, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	chanDelivery, err := chann.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		fmt.Printf("error consuming the queue: %v", err)
		return err
	}

	go func() {
		for msg := range chanDelivery {
			var body T
			json.Unmarshal(msg.Body, &body)
			ack := handler(body)

			if ack == Ack {
				msg.Ack(false)
			}

			if ack == NackRequeue {
				msg.Nack(false, true)
			}

			if ack == NackDiscard {
				msg.Nack(false, false)
			}

			fmt.Printf("Ack type %s\n", ack)
		}
	}()

	return nil

}
