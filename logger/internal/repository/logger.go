package repository

import (
	"github.com/ziliscite/go-micro-logger/internal/data"

	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrDuplicateEntry  = errors.New("duplicate entry")
	ErrInvalidData     = errors.New("invalid data")
	ErrDatabaseTimeout = errors.New("database operation timed out")

	ErrNotFound  = errors.New("entry not found")
	ErrInvalidID = errors.New("invalid ID format")

	ErrNotModified = errors.New("entry not modified")
)

type Repository struct {
	mc *mongo.Collection
}

func New(client *mongo.Client) *Repository {
	return &Repository{
		// Inject the collection dependency
		//
		// If a collection is not available, it will be created
		mc: client.Database("logs").Collection("logs"),
	}
}

func (r Repository) Insert(ctx context.Context, entry *data.Entry) error {
	// Insert a log entry
	res, err := r.mc.InsertOne(ctx, entry)
	if err == nil {
		entry.ID = res.InsertedID.(primitive.ObjectID).Hex()
		return nil
	}

	// Handle specific error types
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, e := range writeErr.WriteErrors {
			switch e.Code {
			case 11000: // Duplicate key error code
				return fmt.Errorf("%w: %v", ErrDuplicateEntry, err)
			case 121: // Document validation error
				return fmt.Errorf("%w: %v", ErrInvalidData, err)
			}
		}
	}

	// Context-related errors
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %v", ErrDatabaseTimeout, err)
	}

	return fmt.Errorf("database error: %w", err)
}

func (r Repository) GetAll(ctx context.Context) ([]data.Entry, error) {
	entries := make([]data.Entry, 0)

	// Get all logs in the collection
	cursor, err := r.mc.Find(ctx, bson.D{}, options.Find().SetSort(
		// Sort by created_at in descending order
		bson.D{{"created_at", -1}},
	))
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	defer func() {
		if err = cursor.Close(ctx); err != nil {
			slog.Error("Failed to close cursor", "error", err)
		}
	}()

	for cursor.Next(ctx) {
		var entry data.Entry
		if err = cursor.Decode(&entry); err != nil {
			return nil, fmt.Errorf("data decoding error: %w", err)
		}

		entries = append(entries, entry)
	}

	// Check for errors
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("database iteration failed: %w", err)
	}

	return entries, nil
}

func (r Repository) Get(ctx context.Context, id string) (*data.Entry, error) {
	// Parse the ID
	entryId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidID, err)
	}

	// Get a log entry
	var entry data.Entry
	err = r.mc.FindOne(ctx, bson.D{{"_id", entryId}}).Decode(&entry)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, fmt.Errorf("%w: id %s", ErrNotFound, id)
		default:
			return nil, fmt.Errorf("database operation failed: %w", err)
		}
	}

	return &entry, nil
}

func (r Repository) Drop(ctx context.Context) error {
	// Drop the collection
	return r.mc.Drop(ctx)
}

func (r Repository) Update(ctx context.Context, entry *data.Entry) (*mongo.UpdateResult, error) {
	// Parse the ID
	entryId, err := primitive.ObjectIDFromHex(entry.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidID, err)
	}

	// Update a log entry
	res, err := r.mc.UpdateByID(ctx, entryId, bson.D{
		{"$set", bson.D{
			// Set the new value for fields
			{"title", entry.Title},
			{"content", entry.Content},
			{"updated_at", time.Now()},
		}},
	})
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, e := range writeErr.WriteErrors {
				switch e.Code {
				case 11000: // Duplicate key
					return nil, fmt.Errorf("%w: %v", ErrDuplicateEntry, err)
				}
			}
		}
		return nil, fmt.Errorf("database update failed: %w", err)
	}

	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("%w: id %s", ErrNotFound, entry.ID)
	}

	if res.ModifiedCount == 0 {
		return nil, fmt.Errorf("%w: id %s", ErrNotModified, entry.ID)
	}

	return res, nil
}
