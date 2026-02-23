package server

import (
	"context"

	contextpb "github.com/blcvn/ba-shared-libs/proto/context"
)

type ContextServer struct {
	contextpb.UnimplementedContextServiceServer
}

func NewContextServer() *ContextServer {
	return &ContextServer{}
}

func (s *ContextServer) SetUserContext(ctx context.Context, req *contextpb.SetContextRequest) (*contextpb.SetContextResponse, error) {
	return &contextpb.SetContextResponse{Success: true}, nil
}

func (s *ContextServer) GetUserContext(ctx context.Context, req *contextpb.GetContextRequest) (*contextpb.GetContextResponse, error) {
	return &contextpb.GetContextResponse{Value: "stub", Found: true}, nil
}
