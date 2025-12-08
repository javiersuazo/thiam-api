package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	exchangeType = "topic"
	contentType  = "application/json"
)

type RabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

func NewRabbitMQPublisher(url, exchange string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQPublisher - dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("RabbitMQPublisher - channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchange,
		exchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("RabbitMQPublisher - declare exchange: %w", err)
	}

	return &RabbitMQPublisher{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
	}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, e event.OutboxEvent) error {
	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
		e.EventType, // routing key (e.g., "user.created")
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  contentType,
			DeliveryMode: amqp.Persistent,
			MessageId:    e.ID.String(),
			Timestamp:    time.Now().UTC(),
			Type:         e.EventType,
			Body:         e.Payload,
		},
	)
}

func (p *RabbitMQPublisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("RabbitMQPublisher - close channel: %w", err)
	}
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("RabbitMQPublisher - close connection: %w", err)
	}
	return nil
}
