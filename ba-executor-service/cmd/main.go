package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/blcvn/backend/services/ba-executor-service/internal/grpc_server"
	"github.com/blcvn/backend/services/ba-executor-service/internal/registry"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// Init Dependencies
	regClient := registry.NewClient("http://tool-registry:8080")

	// Register Services
	grpc_server.RegisterExecutorServer(s, regClient)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
