package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	DURABLE   SimpleQueueType = 0
	TRANSIENT SimpleQueueType = 1
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	dat, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        dat,
	})

	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, SimpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	connChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	queue, err := connChan.QueueDeclare(
		queueName,
		SimpleQueueType == DURABLE,
		SimpleQueueType == TRANSIENT,
		SimpleQueueType == TRANSIENT,
		false,
		nil,
	)

	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = connChan.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return connChan, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	connChan, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return err
	}

	delivery, err := connChan.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	errChan := make(chan error)

	go func() {
		for d := range delivery {
			var t T
			err := json.Unmarshal(d.Body, &t)
			if err != nil {
				errChan <- err
			}

			handler(t)

			err = d.Ack(false)
			if err != nil {
				errChan <- err
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	default:
	}
	return nil
}
