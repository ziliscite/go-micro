package main

import (
	"context"
	"github.com/ziliscite/go-micro-logger/internal/data"
	"github.com/ziliscite/go-micro-logger/internal/repository"
	"log/slog"
	"time"
)

// RPCServer is a specific type before implementing rpc
type RPCServer struct {
	repo *repository.Repository
}

// RPCPayload is the payload we're going to receive from the rpc
type RPCPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// LogInfo log data and write it to mongo
//
// Must start with a capital letter to be exported
// input argument, RPCPayload, must also begin in uppercase to be exported
func (r *RPCServer) LogInfo(payload RPCPayload, res *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	entry := data.Entry{
		Title:     payload.Name,
		Content:   payload.Data,
		CreatedAt: time.Now(),
	}
	err := r.repo.Insert(ctx, &entry)
	if err != nil {
		slog.Error("Failed to insert data", "error", err)
		*res = "Failed to insert data: " + err.Error()
		return err
	}

	*res = "Processed entry: " + entry.ID + " | " + entry.Title
	return nil
}
