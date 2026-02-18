package server

import (
	"context"

	"github.com/blcvn/backend/services/ba-state-service/internal/cache"
	statepb "github.com/blcvn/backend/services/proto/state"
)

type StateServer struct {
	statepb.UnimplementedStateServiceServer
	cache *cache.RedisCache
}

func NewStateServer(c *cache.RedisCache) *StateServer {
	return &StateServer{
		cache: c,
	}
}

func (s *StateServer) SetSessionState(ctx context.Context, req *statepb.SetStateRequest) (*statepb.SetStateResponse, error) {
	err := s.cache.SetState(ctx, req.SessionId, req.Key, req.Value)
	if err != nil {
		return &statepb.SetStateResponse{Success: false}, err
	}
	return &statepb.SetStateResponse{Success: true}, nil
}

func (s *StateServer) GetSessionState(ctx context.Context, req *statepb.GetStateRequest) (*statepb.GetStateResponse, error) {
	val, found, err := s.cache.GetState(ctx, req.SessionId, req.Key)
	if err != nil {
		return nil, err
	}
	return &statepb.GetStateResponse{Value: val, Found: found}, nil
}

func (s *StateServer) ClearSessionState(ctx context.Context, req *statepb.ClearStateRequest) (*statepb.ClearStateResponse, error) {
	err := s.cache.ClearState(ctx, req.SessionId)
	if err != nil {
		return &statepb.ClearStateResponse{Success: false}, err
	}
	return &statepb.ClearStateResponse{Success: true}, nil
}
