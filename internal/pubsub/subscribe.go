package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
				continue
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
				continue
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

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
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

	err = connChan.Qos(10, 0, true)
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
			dat, err := unmarshaller(d.Body)
			if err != nil {
				errChan <- err
				continue
			}

			Acktype := handler(dat)

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
				continue
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
