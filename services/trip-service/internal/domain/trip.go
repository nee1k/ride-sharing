package domain

import (
	"context"
	"ride-sharing/services/trip-service/pkg/types"
)

// TripRepository defines the interface for trip data persistence
type TripRepository interface {
	// Create creates a new trip in the database
	Create(ctx context.Context, trip *types.Trip) error
	
	// GetByID retrieves a trip by its ID
	GetByID(ctx context.Context, id string) (*types.Trip, error)
	
	// Update updates an existing trip
	Update(ctx context.Context, trip *types.Trip) error
	
	// UpdateStatus updates only the status of a trip
	UpdateStatus(ctx context.Context, id string, status types.TripStatus) error
}

// EventPublisher defines the interface for publishing events to RabbitMQ
type EventPublisher interface {
	// PublishTripCreated publishes a trip.event.created event
	PublishTripCreated(ctx context.Context, trip *types.Trip) error
	
	// PublishDriverAssigned publishes a trip.event.driver_assigned event
	PublishDriverAssigned(ctx context.Context, trip *types.Trip) error
	
	// PublishNoDriversFound publishes a trip.event.no_drivers_found event
	PublishNoDriversFound(ctx context.Context, tripID string) error
}

// OSRMClient defines the interface for OSRM routing API
type OSRMClient interface {
	// GetRoute calculates a route between two coordinates
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.Route, error)
}

// FareCalculator defines the interface for calculating trip fares
type FareCalculator interface {
	// CalculateFares calculates fare options for a given route
	CalculateFares(ctx context.Context, route *types.Route) ([]*types.RouteFare, error)
}

// TripService defines the business logic interface for trip operations
type TripService interface {
	// PreviewTrip calculates route and fare options without creating a trip
	PreviewTrip(ctx context.Context, userID string, pickup, destination *types.Coordinate) (*types.Route, []*types.RouteFare, error)
	
	// CreateTrip creates a new trip with the selected fare
	CreateTrip(ctx context.Context, userID string, fareID string, pickup, destination *types.Coordinate) (*types.Trip, error)
	
	// HandleDriverResponse processes a driver's accept/decline response
	HandleDriverResponse(ctx context.Context, tripID string, driverID string, accepted bool) error
}
