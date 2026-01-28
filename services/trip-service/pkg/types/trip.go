package types

import "time"

// TripStatus represents the current status of a trip
type TripStatus string

const (
	TripStatusPending      TripStatus = "pending"
	TripStatusCreated      TripStatus = "created"
	TripStatusDriverFound  TripStatus = "driver_found"
	TripStatusDriverAssigned TripStatus = "driver_assigned"
	TripStatusInProgress   TripStatus = "in_progress"
	TripStatusCompleted    TripStatus = "completed"
	TripStatusCancelled    TripStatus = "cancelled"
)

// CarPackageSlug represents the type of vehicle package
type CarPackageSlug string

const (
	CarPackageSedan   CarPackageSlug = "sedan"
	CarPackageSUV     CarPackageSlug = "suv"
	CarPackageVAN    CarPackageSlug = "van"
	CarPackageLuxury CarPackageSlug = "luxury"
)

// Trip represents a ride-sharing trip
type Trip struct {
	ID          string      `json:"id" bson:"_id"`
	UserID      string      `json:"userID" bson:"user_id"`
	Status      TripStatus  `json:"status" bson:"status"`
	Route       *Route      `json:"route" bson:"route"`
	SelectedFare *RouteFare `json:"selectedFare,omitempty" bson:"selected_fare,omitempty"`
	Driver      *Driver     `json:"driver,omitempty" bson:"driver,omitempty"`
	CreatedAt   time.Time   `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time   `json:"updatedAt" bson:"updated_at"`
}

// Route represents the route information for a trip
type Route struct {
	Distance float64      `json:"distance" bson:"distance"` // in meters
	Duration float64      `json:"duration" bson:"duration"` // in seconds
	Geometry []*Geometry  `json:"geometry" bson:"geometry"`
}

// Geometry represents a geometry segment of the route
type Geometry struct {
	Coordinates []*Coordinate `json:"coordinates" bson:"coordinates"`
}

// Coordinate represents a geographic coordinate
type Coordinate struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

// RouteFare represents pricing information for a route
type RouteFare struct {
	ID              string        `json:"id" bson:"_id"`
	PackageSlug     CarPackageSlug `json:"packageSlug" bson:"package_slug"`
	BasePrice       float64       `json:"basePrice" bson:"base_price"`
	TotalPriceInCents int64        `json:"totalPriceInCents,omitempty" bson:"total_price_in_cents,omitempty"`
	ExpiresAt       time.Time     `json:"expiresAt" bson:"expires_at"`
	Route           *Route        `json:"route" bson:"route"`
}

// Driver represents a driver assigned to a trip
type Driver struct {
	ID           string      `json:"id" bson:"_id"`
	Name         string      `json:"name" bson:"name"`
	Location     *Coordinate `json:"location" bson:"location"`
	Geohash      string      `json:"geohash" bson:"geohash"`
	ProfilePicture string    `json:"profilePicture" bson:"profile_picture"`
	CarPlate     string      `json:"carPlate" bson:"car_plate"`
}
