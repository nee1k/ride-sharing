package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/pkg/types"
)

// MongoTripRepository implements TripRepository using MongoDB
type MongoTripRepository struct {
	collection *mongo.Collection
}

// NewMongoTripRepository creates a new MongoDB trip repository
func NewMongoTripRepository(db *mongo.Database) domain.TripRepository {
	collection := db.Collection("trips")
	
	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}
	
	_, _ = collection.Indexes().CreateMany(context.Background(), indexes)
	
	return &MongoTripRepository{
		collection: collection,
	}
}

// Create creates a new trip in the database
func (r *MongoTripRepository) Create(ctx context.Context, trip *types.Trip) error {
	now := time.Now()
	trip.CreatedAt = now
	trip.UpdatedAt = now
	
	_, err := r.collection.InsertOne(ctx, trip)
	return err
}

// GetByID retrieves a trip by its ID
func (r *MongoTripRepository) GetByID(ctx context.Context, id string) (*types.Trip, error) {
	var trip types.Trip
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&trip)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &trip, nil
}

// Update updates an existing trip
func (r *MongoTripRepository) Update(ctx context.Context, trip *types.Trip) error {
	trip.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": trip.ID}
	update := bson.M{"$set": trip}
	
	opts := options.Update().SetUpsert(false)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// UpdateStatus updates only the status of a trip
func (r *MongoTripRepository) UpdateStatus(ctx context.Context, id string, status types.TripStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
