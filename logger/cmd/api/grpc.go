package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ziliscite/go-micro-logger/internal/data"
	genproto "github.com/ziliscite/go-micro-logger/proto/genproto"
	"time"

	"github.com/ziliscite/go-micro-logger/internal/repository"
)

type LogServer struct {
	// this finna be required for every grpc service, ever
	// to ensure backward compatibility
	genproto.UnimplementedLogServiceServer // If you hover over this, you can see the methods, one of which
	// is `WriteLog(ctx context.Context, req *genproto.LogRequest) (*genproto.LogResponse, error)`
	//
	// To that, we should also define a method to match that signature to actually write the log

	repo *repository.Repository
}

// WriteLog uses the same format as the proto file, but now we added context as request and error as return value.
// well, the context and the error are required for every grpc service -- given that `genproto.UnimplementedLogServiceServer`
// have the same method with the same signature. I guess it's just the way it is.
func (l *LogServer) WriteLog(ctx context.Context, req *genproto.LogRequest) (*genproto.LogResponse, error) {
	// Get the log entry [entry here the method signature in proto file]
	input := req.GetEntry()

	var msg string
	err := l.repo.Insert(ctx, &data.Entry{
		Title:     input.Name,
		Content:   input.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEntry):
			msg = repository.ErrDuplicateEntry.Error()
		case errors.Is(err, repository.ErrInvalidData):
			msg = repository.ErrInvalidData.Error()
		case errors.Is(err, repository.ErrDatabaseTimeout):
			msg = repository.ErrDatabaseTimeout.Error()
		default:
			msg = "Something went wrong"
		}
	} else {
		msg = fmt.Sprintf("%s is successfully Logged!", input.Name)
	}

	// weird error handling...

	return &genproto.LogResponse{Response: msg}, err
}
