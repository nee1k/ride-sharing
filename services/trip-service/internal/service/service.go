package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/pkg/types"
)

// TripServiceImpl implements the TripService interface
type TripServiceImpl struct {
	repo          domain.TripRepository
	osrmClient    domain.OSRMClient
	fareCalculator domain.FareCalculator
	eventPublisher domain.EventPublisher
}

// NewTripService creates a new trip service
func NewTripService(
	repo domain.TripRepository,
	osrmClient domain.OSRMClient,
	fareCalculator domain.FareCalculator,
	eventPublisher domain.EventPublisher,
) domain.TripService {
	return &TripServiceImpl{
		repo:           repo,
		osrmClient:     osrmClient,
		fareCalculator: fareCalculator,
		eventPublisher: eventPublisher,
	}
}

// PreviewTrip calculates route and fare options without creating a trip
func (s *TripServiceImpl) PreviewTrip(ctx context.Context, userID string, pickup, destination *types.Coordinate) (*types.Route, []*types.RouteFare, error) {
	// Get route from OSRM
	route, err := s.osrmClient.GetRoute(ctx, pickup, destination)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get route: %w", err)
	}

	// Calculate fares
	fares, err := s.fareCalculator.CalculateFares(ctx, route)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate fares: %w", err)
	}

	return route, fares, nil
}

// CreateTrip creates a new trip with the selected fare
func (s *TripServiceImpl) CreateTrip(ctx context.Context, userID string, fareID string, pickup, destination *types.Coordinate) (*types.Trip, error) {
	// First, get the route and fares to find the selected fare
	route, fares, err := s.PreviewTrip(ctx, userID, pickup, destination)
	if err != nil {
		return nil, err
	}

	// Find the selected fare
	var selectedFare *types.RouteFare
	for _, fare := range fares {
		if fare.ID == fareID {
			selectedFare = fare
			break
		}
	}

	if selectedFare == nil {
		return nil, fmt.Errorf("fare not found: %s", fareID)
	}

	// Check if fare has expired
	if time.Now().After(selectedFare.ExpiresAt) {
		return nil, fmt.Errorf("fare has expired")
	}

	// Create trip
	trip := &types.Trip{
		ID:          uuid.New().String(),
		UserID:      userID,
		Status:      types.TripStatusCreated,
		Route:       route,
		SelectedFare: selectedFare,
	}

	// Save to database
	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	// Publish trip created event
	if err := s.eventPublisher.PublishTripCreated(ctx, trip); err != nil {
		// Log error but don't fail the trip creation
		fmt.Printf("Warning: failed to publish trip created event: %v\n", err)
	}

	return trip, nil
}

// HandleDriverResponse processes a driver's accept/decline response
func (s *TripServiceImpl) HandleDriverResponse(ctx context.Context, tripID string, driverID string, accepted bool) error {
	// Get trip from database
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("failed to get trip: %w", err)
	}

	if trip == nil {
		return fmt.Errorf("trip not found: %s", tripID)
	}

	if accepted {
		// Update trip status to driver assigned
		if err := s.repo.UpdateStatus(ctx, tripID, types.TripStatusDriverAssigned); err != nil {
			return fmt.Errorf("failed to update trip status: %w", err)
		}

		// Get updated trip
		trip, err = s.repo.GetByID(ctx, tripID)
		if err != nil {
			return fmt.Errorf("failed to get updated trip: %w", err)
		}

		// Publish driver assigned event
		if err := s.eventPublisher.PublishDriverAssigned(ctx, trip); err != nil {
			return fmt.Errorf("failed to publish driver assigned event: %w", err)
		}
	} else {
		// Driver declined - for now, we'll just log it
		// In a real system, we might want to find another driver or mark the trip as needing a new driver
		fmt.Printf("Driver %s declined trip %s\n", driverID, tripID)
	}

	return nil
}
