package main

import (
	"fmt"
	"log"
	"net"

	"github.com/blcvn/backend/services/ba-context-service/internal/server"
	contextpb "github.com/blcvn/ba-shared-libs/proto/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50055)) // Port 50055
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	contextServer := server.NewContextServer()
	contextpb.RegisterContextServiceServer(s, contextServer)

	// Enable reflection
	reflection.Register(s)

	log.Printf("Context Service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
