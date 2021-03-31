package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	pb "github.com/leplasmo/chaman/registration/proto"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
        
        
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = "0.0.0.0:50051"
)

var collection *mongo.Collection

type server struct{}

type registrationDocument struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Email string            `bson:"email"`
}

func (*server) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	// email := req.GetEmail()

	res := &pb.RegisterUserResponse{
		Id:     "1",
		Status: pb.StatusCode_SUCCESS,
	}

	return res, nil
}

func main() {
	// dump file name and line number if program crashes
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("\n\nStarting Registration service...")

	uri := "mongodb://admin:admin@localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Println("Failed to connect to the database server")
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Println("Failed to disconnect properly from database")
			panic(err)

		}
	}()

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Println("Failed to ping the database server")
		panic(err)
	}
	fmt.Println("Successfully connected to the database")

	collection = client.Database("registrationService").Collection("registrations")

	// open a listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Printf("Successfully opened the listener port on: %s", port)

	// set the grpc server options
	opts := []grpc.ServerOption{}
	tls := false
	if tls {
		certFile := "ssl/tls.crt"
		keyFile := "ssl/tls.key"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed to load certificates for TLS: %v", err)
			return
		}
		opts = append(opts, grpc.Creds(creds))
		fmt.Printf("Succesfully loaded the TLS certificates")

	}

	// create the grpc server
	s := grpc.NewServer(opts...)
	pb.RegisterRegistrationServiceServer(s, &server{})

	// launch the server in a goroutine
	go func() {

		if err = s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// listen for interrupts - graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch // block waiting for interrupt

	fmt.Println("\nReceived interrupt. Stopping the service...")
	s.Stop()
	fmt.Println("Closing the listener...")
	lis.Close()
	fmt.Println("\n\nRegistration service is stopped")
}
