package main

import (
	"fmt"
	"github.com/ziliscite/go-micro-logger/internal/repository"
	genproto "github.com/ziliscite/go-micro-logger/proto/genproto"
	"google.golang.org/grpc"
	"net"
	"net/rpc"

	"context"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ApiPort  = "80"
	RPCPort  = "5001"
	GRPCPort = "50001"
)

type application struct {
	repo *repository.Repository
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := openMongo(ctx)
	if err != nil {
		slog.Error("Failed to connect to mongo", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			slog.Error("Failed to disconnect from mongo", "error", err)
			os.Exit(1)
		}
	}()

	app := application{
		repo: repository.New(client),
	}

	// Register rpc -- must be a pointer
	if err = rpc.Register(&RPCServer{
		repo: app.repo,
	}); err != nil {
		slog.Error("Failed to register rpc", "error", err.Error())
		os.Exit(1)
	}

	go app.rpcListen()

	go app.grpcListen()

	if err = app.serve(); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func (app *application) rpcListen() {
	slog.Info("Starting rpc server", "port", RPCPort)

	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", RPCPort))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func (app *application) grpcListen() {
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", GRPCPort))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer listen.Close()

	srv := grpc.NewServer()

	// Register the service
	genproto.RegisterLogServiceServer(srv, &LogServer{
		repo: app.repo,
	})

	slog.Info("Starting grpc server", "port", GRPCPort)

	if err = srv.Serve(listen); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func openMongo(ctx context.Context) (*mongo.Client, error) {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	opts := options.Client().ApplyURI(os.Getenv("MONGO_URL"))
	opts.SetAuth(options.Credential{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	})

	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Send a ping to confirm a successful connection
	if err = client.Database(os.Getenv("MONGO_DATABASE")).RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
