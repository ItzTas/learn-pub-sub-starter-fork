package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	DURABLE SimpleQueueType = iota
	TRANSIENT
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

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
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
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
	handler func(T) Acktype,
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

			Acktype := handler(t)

			switch Acktype {
			case Ack:
				err = d.Ack(false)
				log.Println("Acked the message")
			case NackRequeue:
				err = d.Nack(false, true)
				log.Println("NACKed the message for requeue")
			case NackDiscard:
				err = d.Nack(false, false)
				log.Println("NACKed the message to discard")
			}

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
