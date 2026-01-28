package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/contracts"
)

// EventConsumer handles consuming events from RabbitMQ
type EventConsumer struct {
	channel  *amqp.Channel
	service  *service.TripServiceImpl
}

// NewEventConsumer creates a new event consumer
func NewEventConsumer(ch *amqp.Channel, tripService *service.TripServiceImpl) (*EventConsumer, error) {
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

	return &EventConsumer{
		channel: ch,
		service: tripService,
	}, nil
}

// StartDriverResponseConsumer starts consuming driver response messages
func (c *EventConsumer) StartDriverResponseConsumer(ctx context.Context) error {
	// Declare queue
	queue, err := c.channel.QueueDeclare(
		"driver_trip_response", // name
		true,                    // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = c.channel.QueueBind(
		queue.Name,              // queue name
		contracts.DriverCmdTripAccept, // routing key (also handles decline)
		"trip_exchange",         // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Also bind for decline
	err = c.channel.QueueBind(
		queue.Name,
		contracts.DriverCmdTripDecline,
		"trip_exchange",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue for decline: %w", err)
	}

	msgs, err := c.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.handleDriverResponse(ctx, msg)
			}
		}
	}()

	log.Println("Started driver response consumer")
	return nil
}

// DriverResponseMessage represents the driver response message
type DriverResponseMessage struct {
	TripID   string `json:"tripID"`
	RiderID  string `json:"riderID"`
	DriverID string `json:"driverID"`
	Accepted bool   `json:"accepted"`
}

func (c *EventConsumer) handleDriverResponse(ctx context.Context, msg amqp.Delivery) {
	var response DriverResponseMessage
	if err := json.Unmarshal(msg.Body, &response); err != nil {
		log.Printf("Failed to unmarshal driver response: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("Received driver response: tripID=%s, driverID=%s, accepted=%v",
		response.TripID, response.DriverID, response.Accepted)

	if err := c.service.HandleDriverResponse(ctx, response.TripID, response.DriverID, response.Accepted); err != nil {
		log.Printf("Failed to handle driver response: %v", err)
		msg.Nack(false, true) // Requeue on error
		return
	}

	msg.Ack(false)
}
