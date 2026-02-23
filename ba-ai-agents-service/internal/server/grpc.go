package server

import (
	"context"

	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/handler"
	"github.com/blcvn/backend/services/ba-ai-agents-service/internal/models"
	aiagentspb "github.com/blcvn/ba-shared-libs/proto/ai_agents"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AIAgentsServer struct {
	aiagentspb.UnimplementedAIAgentsServiceServer
	svc *handler.AIAgentsService
}

func NewAIAgentsServer(svc *handler.AIAgentsService) *AIAgentsServer {
	return &AIAgentsServer{svc: svc}
}

func (s *AIAgentsServer) GetAgent(ctx context.Context, req *aiagentspb.GetAgentRequest) (*aiagentspb.GetAgentResponse, error) {
	persona, err := s.svc.GetPersona(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "agent not found: %v", err)
	}
	return &aiagentspb.GetAgentResponse{
		Agent: mapPersonaToProto(persona),
	}, nil
}

func (s *AIAgentsServer) ListAgents(ctx context.Context, req *aiagentspb.ListAgentsRequest) (*aiagentspb.ListAgentsResponse, error) {
	personas, err := s.svc.ListPersonas()
	if err != nil {
		return nil, err
	}
	var agents []*aiagentspb.Agent
	for _, p := range personas {
		agents = append(agents, mapPersonaToProto(p))
	}
	return &aiagentspb.ListAgentsResponse{Agents: agents}, nil
}

// Stub Skills for now
func (s *AIAgentsServer) GetSkill(ctx context.Context, req *aiagentspb.GetSkillRequest) (*aiagentspb.GetSkillResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *AIAgentsServer) ListSkills(ctx context.Context, req *aiagentspb.ListSkillsRequest) (*aiagentspb.ListSkillsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func mapPersonaToProto(p *models.AgentPersona) *aiagentspb.Agent {
	return &aiagentspb.Agent{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Persona:     p.Persona,
		IsActive:    p.IsActive,
		Skills:      p.Skills,
	}
}
