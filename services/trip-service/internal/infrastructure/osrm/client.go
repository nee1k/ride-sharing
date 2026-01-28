package osrm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/env"
)

// OSRMClient implements the OSRMClient interface for route calculation
type OSRMClient struct {
	baseURL string
	client  *http.Client
}

// NewOSRMClient creates a new OSRM client
func NewOSRMClient() domain.OSRMClient {
	baseURL := env.GetString("OSRM_URL", "http://router.project-osrm.org")
	
	return &OSRMClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// OSRMResponse represents the response from OSRM API
type OSRMResponse struct {
	Code   string      `json:"code"`
	Routes []OSRMRoute `json:"routes"`
}

// OSRMRoute represents a route in OSRM response
type OSRMRoute struct {
	Distance float64            `json:"distance"` // in meters
	Duration float64            `json:"duration"` // in seconds
	Geometry OSRMGeometry       `json:"geometry"` // GeoJSON geometry
}

// OSRMGeometry represents the GeoJSON geometry from OSRM
type OSRMGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"` // [lon, lat] pairs
}

// GetRoute calculates a route between two coordinates
func (c *OSRMClient) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.Route, error) {
	// OSRM API format: /route/v1/driving/{lon1},{lat1};{lon2},{lat2}
	endpoint := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f",
		c.baseURL,
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Request geometry as encoded polyline
	q := url.Values{}
	q.Set("overview", "full")
	q.Set("geometries", "geojson")
	req.URL.RawQuery = q.Encode()
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OSRM API error: status %d, body: %s", resp.StatusCode, string(body))
	}
	
	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found: code %s", osrmResp.Code)
	}
	
	route := osrmResp.Routes[0]
	
	// Convert coordinates from OSRM format [lon, lat] to our format
	coordinates := make([]*types.Coordinate, 0, len(route.Geometry.Coordinates))
	for _, coord := range route.Geometry.Coordinates {
		if len(coord) >= 2 {
			coordinates = append(coordinates, &types.Coordinate{
				Latitude:  coord[1], // OSRM uses [lon, lat], we need lat first
				Longitude: coord[0],
			})
		}
	}
	
	return &types.Route{
		Distance: route.Distance,
		Duration: route.Duration,
		Geometry: []*types.Geometry{
			{
				Coordinates: coordinates,
			},
		},
	}, nil
}
