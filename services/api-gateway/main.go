package main

import (
	"encoding/json"
	"log"
	"net/http"

	"ride-sharing/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
)

func main() {
	log.Println("🚀 Starting API Gateway on http://localhost:8081")
	log.Println("✅ Server ready! Try: curl http://localhost:8081")

	// Health check
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("✅ API Gateway is running!"))
	})

	// Trip Preview endpoint - this is what the frontend calls when you click on the map
	http.HandleFunc("/trip/preview", corsHandler(previewTripHandler))
	
	// Trip Start endpoint - this is what the frontend calls when you select a fare
	http.HandleFunc("/trip/start", corsHandler(startTripHandler))

	log.Println("📡 Listening on", httpAddr)
	log.Println("🌐 Frontend should connect to: http://localhost:8081")
	http.ListenAndServe(httpAddr, nil)
}

// CORS handler - allows frontend (running on different port) to call this API
func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// PreviewTripHandler - handles trip preview requests from frontend
func previewTripHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req struct {
		UserID      string `json:"userID"`
		Pickup      struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"pickup"`
		Destination struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"destination"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Error parsing request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("📍 Preview Trip Request:")
	log.Printf("   User: %s", req.UserID)
	log.Printf("   Pickup: %.4f, %.4f", req.Pickup.Latitude, req.Pickup.Longitude)
	log.Printf("   Destination: %.4f, %.4f", req.Destination.Latitude, req.Destination.Longitude)

	// For now, return mock data so you can see it work!
	// Later, we'll connect this to the Trip Service
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"route": map[string]interface{}{
				"distance": 5000.0,  // 5km in meters
				"duration": 600.0,   // 10 minutes in seconds
				"geometry": []map[string]interface{}{
					{
						"coordinates": []map[string]float64{
							{"latitude": req.Pickup.Latitude, "longitude": req.Pickup.Longitude},
							{"latitude": req.Destination.Latitude, "longitude": req.Destination.Longitude},
						},
					},
				},
			},
			"rideFares": []map[string]interface{}{
				{
					"id":              "fare-sedan-1",
					"packageSlug":     "sedan",
					"basePrice":       12.50,
					"totalPriceInCents": 1250,
					"expiresAt":       "2026-01-28T12:00:00Z",
					"route": map[string]interface{}{
						"distance": 5000.0,
						"duration": 600.0,
						"geometry": []map[string]interface{}{},
					},
				},
				{
					"id":              "fare-suv-1",
					"packageSlug":     "suv",
					"basePrice":       15.00,
					"totalPriceInCents": 1500,
					"expiresAt":       "2026-01-28T12:00:00Z",
					"route": map[string]interface{}{
						"distance": 5000.0,
						"duration": 600.0,
						"geometry": []map[string]interface{}{},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("✅ Preview trip response sent!")
}

// StartTripHandler - handles trip creation requests
func startTripHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RideFareID string `json:"rideFareID"`
		UserID     string `json:"userID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Error parsing request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("🚗 Start Trip Request:")
	log.Printf("   User: %s", req.UserID)
	log.Printf("   Fare ID: %s", req.RideFareID)

	// Mock response - later we'll connect to Trip Service
	response := map[string]interface{}{
		"tripID": "trip-" + req.RideFareID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Println("✅ Trip started!")
}