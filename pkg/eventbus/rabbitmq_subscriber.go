package eventbus

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQSubscriber struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

func NewRabbitMQSubscriber(url, exchange, queue string) (*RabbitMQSubscriber, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQSubscriber - dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf("RabbitMQSubscriber - channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()

		return nil, fmt.Errorf("RabbitMQSubscriber - declare exchange: %w", err)
	}

	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()

		return nil, fmt.Errorf("RabbitMQSubscriber - declare queue: %w", err)
	}

	return &RabbitMQSubscriber{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (s *RabbitMQSubscriber) Subscribe(ctx context.Context, topic string) (<-chan Event, error) {
	routingKey := topic + ".#"

	err := s.channel.QueueBind(
		s.queue,
		routingKey,
		s.exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQSubscriber - bind queue: %w", err)
	}

	msgs, err := s.channel.Consume(
		s.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQSubscriber - consume: %w", err)
	}

	events := make(chan Event)

	go func() {
		defer close(events)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				events <- Event{
					ID:      msg.MessageId,
					Type:    msg.Type,
					Payload: msg.Body,
				}

				//nolint:errcheck // best effort ack - if it fails, message will be redelivered
				msg.Ack(false)
			}
		}
	}()

	return events, nil
}

func (s *RabbitMQSubscriber) Close() error {
	if err := s.channel.Close(); err != nil {
		return fmt.Errorf("RabbitMQSubscriber - close channel: %w", err)
	}

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("RabbitMQSubscriber - close connection: %w", err)
	}

	return nil
}
