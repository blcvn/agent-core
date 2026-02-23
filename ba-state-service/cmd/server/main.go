package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/blcvn/backend/services/ba-state-service/internal/cache"
	"github.com/blcvn/backend/services/ba-state-service/internal/server"
	statepb "github.com/blcvn/ba-shared-libs/proto/state"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Initialize Cache (Redis)
	redisCache := cache.NewRedisCache(redisAddr, "", 0, 24*time.Hour)
	defer redisCache.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50054)) // Port 50054
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	stateServer := server.NewStateServer(redisCache)
	statepb.RegisterStateServiceServer(s, stateServer)

	// Enable reflection
	reflection.Register(s)

	log.Printf("State Service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
