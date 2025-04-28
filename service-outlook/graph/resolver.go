package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	mongoClient *mongo.Client
	initOnce    sync.Once
}

// MongoDB Document Structures
type MongoEvent struct {
	EventID             string                         `bson:"eventID"`
	Subject             string                         `bson:"subject"`
	Start               time.Time                      `bson:"start"`
	End                 time.Time                      `bson:"end"`
	DurationMinutes     int32                          `bson:"durationMinutes"`
	Rooms               []MongoRoom                    `bson:"rooms"`
	DepartmentBreakdown []MongoDepartmentParticipation `bson:"department_breakdown"`
}

type MongoRoom struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
}

type MongoDepartmentParticipation struct {
	Department    string `bson:"department"`
	AttendeeCount int32  `bson:"attendee_count"`
	IsExternal    bool   `bson:"is_external"`
}

func (r *Resolver) initMongoDB() error {
	var initErr error
	r.initOnce.Do(func() {
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			initErr = fmt.Errorf("MONGODB_URI environment variable not set")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			initErr = fmt.Errorf("failed to connect to MongoDB: %v", err)
			return
		}

		// Verify the connection
		err = client.Ping(ctx, nil)
		if err != nil {
			initErr = fmt.Errorf("failed to ping MongoDB: %v", err)
			return
		}

		r.mongoClient = client
		log.Println("Successfully connected to MongoDB")
	})
	return initErr
}

func (r *Resolver) getCollection() (*mongo.Collection, error) {
	if err := r.initMongoDB(); err != nil {
		return nil, fmt.Errorf("MongoDB initialization error: %v", err)
	}
	return r.mongoClient.Database("bookings_db").Collection("events"), nil
}
