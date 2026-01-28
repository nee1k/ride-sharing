package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/contracts"
)

// EventPublisher implements the EventPublisher interface
type EventPublisher struct {
	channel *amqp.Channel
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(ch *amqp.Channel) (domain.EventPublisher, error) {
	// Declare trip exchange
	err := ch.ExchangeDeclare(
		"trip_exchange", // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare trip exchange: %w", err)
	}

	return &EventPublisher{
		channel: ch,
	}, nil
}

// PublishTripCreated publishes a trip.event.created event
func (p *EventPublisher) PublishTripCreated(ctx context.Context, trip *types.Trip) error {
	return p.publishEvent(ctx, contracts.TripEventCreated, trip)
}

// PublishDriverAssigned publishes a trip.event.driver_assigned event
func (p *EventPublisher) PublishDriverAssigned(ctx context.Context, trip *types.Trip) error {
	return p.publishEvent(ctx, contracts.TripEventDriverAssigned, trip)
}

// PublishNoDriversFound publishes a trip.event.no_drivers_found event
func (p *EventPublisher) PublishNoDriversFound(ctx context.Context, tripID string) error {
	data := map[string]string{
		"trip_id": tripID,
	}
	return p.publishEvent(ctx, contracts.TripEventNoDriversFound, data)
}

// publishEvent is a helper method to publish events to RabbitMQ
func (p *EventPublisher) publishEvent(ctx context.Context, routingKey string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	err = p.channel.PublishWithContext(
		ctx,
		"trip_exchange", // exchange
		routingKey,      // routing key
		false,           // mandatory
		false,           // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published event: %s", routingKey)
	return nil
}
