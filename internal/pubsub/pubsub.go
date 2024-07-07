package pubsub

import (
	"bytes"
	"encoding/gob"

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

func DecodeGob[T any](data []byte) (T, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	var t T
	err := decoder.Decode(&t)
	return t, err
}
