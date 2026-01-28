package osrm

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/pkg/types"
	"time"

	"github.com/google/uuid"
)

// FareCalculator implements fare calculation logic
type FareCalculator struct {
	basePrices map[types.CarPackageSlug]float64
}

// NewFareCalculator creates a new fare calculator
func NewFareCalculator() domain.FareCalculator {
	return &FareCalculator{
		basePrices: map[types.CarPackageSlug]float64{
			types.CarPackageSedan:   2.50,  // $2.50 per km
			types.CarPackageSUV:     3.00,  // $3.00 per km
			types.CarPackageVAN:    3.50,  // $3.50 per km
			types.CarPackageLuxury: 5.00,  // $5.00 per km
		},
	}
}

// CalculateFares calculates fare options for a given route
func (fc *FareCalculator) CalculateFares(ctx context.Context, route *types.Route) ([]*types.RouteFare, error) {
	// Convert distance from meters to kilometers
	distanceKm := route.Distance / 1000.0
	
	fares := make([]*types.RouteFare, 0, len(fc.basePrices))
	expiresAt := time.Now().Add(5 * time.Minute) // Fares expire in 5 minutes
	
	for packageSlug, basePricePerKm := range fc.basePrices {
		basePrice := basePricePerKm * distanceKm
		
		// Add base fee
		basePrice += 2.0 // $2.00 base fee
		
		// Convert to cents
		totalPriceInCents := int64(basePrice * 100)
		
		fare := &types.RouteFare{
			ID:                uuid.New().String(),
			PackageSlug:       packageSlug,
			BasePrice:         basePrice,
			TotalPriceInCents: totalPriceInCents,
			ExpiresAt:         expiresAt,
			Route:             route,
		}
		
		fares = append(fares, fare)
	}
	
	return fares, nil
}
