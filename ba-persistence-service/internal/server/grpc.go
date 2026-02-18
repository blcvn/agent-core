package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/approval"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/document"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/graph"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/review"
	"github.com/blcvn/backend/services/ba-persistence-service/internal/repository/task"

	v32 "github.com/blcvn/backend/services/ba-agent-service/domain/v3.2"
	persistencepb "github.com/blcvn/backend/services/proto/persistence"
	baagent "github.com/blcvn/kratos-proto/go/ba-agent"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PersistenceServer implements the gRPC server for persistence
type PersistenceServer struct {
	persistencepb.UnimplementedPersistenceServiceServer
	taskRepo     *task.TaskRepository
	docRepo      *document.DocumentRepository
	approvalRepo *approval.ApprovalRepository
	graphRepo    *graph.GraphRepository
	reviewRepo   *review.ReviewRepository
}

// NewPersistenceServer creates a new instance of PersistenceServer
func NewPersistenceServer(
	taskRepo *task.TaskRepository,
	docRepo *document.DocumentRepository,
	approvalRepo *approval.ApprovalRepository,
	graphRepo *graph.GraphRepository,
	reviewRepo *review.ReviewRepository,
) *PersistenceServer {
	return &PersistenceServer{
		taskRepo:     taskRepo,
		docRepo:      docRepo,
		approvalRepo: approvalRepo,
		graphRepo:    graphRepo,
		reviewRepo:   reviewRepo,
	}
}

// CreateTask saves a new task
func (s *PersistenceServer) CreateTask(ctx context.Context, req *persistencepb.CreateTaskRequest) (*persistencepb.CreateTaskResponse, error) {
	// ... (Implementation unchanged)
	status := baagent.TaskStatus_TASK_PENDING // Default
	if req.Status != "" {
		if val, ok := baagent.TaskStatus_value[req.Status]; ok {
			status = baagent.TaskStatus(val)
		}
	}

	task := &baagent.AgentTask{
		Id:     req.TaskId,
		Status: status,
	}

	err := s.taskRepo.CreateTask(ctx, task, req.WorkflowMode)
	if err != nil {
		return &persistencepb.CreateTaskResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &persistencepb.CreateTaskResponse{
		Success: true,
	}, nil
}

func (s *PersistenceServer) GetTask(ctx context.Context, req *persistencepb.GetTaskRequest) (*persistencepb.GetTaskResponse, error) {
	agentTask, mode, err := s.taskRepo.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	resultData, _ := json.Marshal(agentTask)
	return &persistencepb.GetTaskResponse{
		TaskId:         agentTask.Id,
		Status:         agentTask.Status.String(),
		WorkflowMode:   mode,
		ResultDataJson: resultData,
		CreatedAt:      timestamppb.New(time.Now()),
		UpdatedAt:      timestamppb.New(time.Now()),
	}, nil
}

func (s *PersistenceServer) UpdateTask(ctx context.Context, req *persistencepb.UpdateTaskRequest) (*persistencepb.UpdateTaskResponse, error) {
	var status baagent.TaskStatus
	if val, ok := baagent.TaskStatus_value[req.Status]; ok {
		status = baagent.TaskStatus(val)
	}
	task := &baagent.AgentTask{
		Id:     req.TaskId,
		Status: status,
	}
	err := s.taskRepo.UpdateTask(ctx, task)
	if err != nil {
		return &persistencepb.UpdateTaskResponse{Success: false, Error: err.Error()}, nil
	}
	return &persistencepb.UpdateTaskResponse{Success: true}, nil
}

func (s *PersistenceServer) ListTasks(ctx context.Context, req *persistencepb.ListTasksRequest) (*persistencepb.ListTasksResponse, error) {
	return &persistencepb.ListTasksResponse{}, nil // Stub
}

// Document Implementation
func (s *PersistenceServer) CreateDocument(ctx context.Context, req *persistencepb.CreateDocumentRequest) (*persistencepb.CreateDocumentResponse, error) {
	// Map Tier string to enum
	var tier v32.RequirementTier
	switch req.Tier {
	case "PRD":
		tier = v32.TierPRD
	case "URD_INDEX":
		tier = v32.TierURDIndex
	case "URD_OUTLINE":
		tier = v32.TierURDOutline
	case "URD_FULL":
		tier = v32.TierURDFull
	default:
		tier = v32.TierUnspecified
	}

	doc := &v32.Document{
		ID:               req.Id,
		ProjectID:        req.ProjectId,
		ParentDocumentID: req.ParentId,
		RootDocumentID:   req.RootId,
		Tier:             tier,
		ModuleName:       req.ModuleName,
		Version:          int(req.Version),
		Content:          req.Content,
		Status:           v32.DocumentStatus(req.Status),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if doc.Status == "" {
		doc.Status = v32.DocumentStatusDraft
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return &persistencepb.CreateDocumentResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &persistencepb.CreateDocumentResponse{
		Success: true,
	}, nil
}

func (s *PersistenceServer) GetDocument(ctx context.Context, req *persistencepb.GetDocumentRequest) (*persistencepb.GetDocumentResponse, error) {
	doc, err := s.docRepo.Get(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &persistencepb.GetDocumentResponse{
		Id:      doc.ID,
		Title:   doc.ModuleName, // Mapping ModuleName to Title for now
		Content: doc.Content,
	}, nil
}

// Graph Stub
func (s *PersistenceServer) ExecuteGraphQuery(ctx context.Context, req *persistencepb.GraphQueryRequest) (*persistencepb.GraphQueryResponse, error) {
	return &persistencepb.GraphQueryResponse{ResultJson: "{}"}, nil
}

// Approval Implementation
func (s *PersistenceServer) CreateApproval(ctx context.Context, req *persistencepb.CreateApprovalRequest) (*persistencepb.CreateApprovalResponse, error) {
	// First fetch document to get Tier
	doc, err := s.docRepo.Get(ctx, req.DocumentId)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	approval := &v32.Approval{
		ID:         fmt.Sprintf("app_%d", time.Now().UnixNano()), // Generate ID
		DocumentID: req.DocumentId,
		Tier:       doc.Tier,
		Comment:    "Approved via Persistence Service",
		ApprovedBy: req.Approver,
		ApprovedAt: time.Now(),
	}

	if err := s.approvalRepo.Create(ctx, approval); err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	return &persistencepb.CreateApprovalResponse{
		Success: true,
	}, nil
}

// Review Stub
func (s *PersistenceServer) CreateReview(ctx context.Context, req *persistencepb.CreateReviewRequest) (*persistencepb.CreateReviewResponse, error) {
	return &persistencepb.CreateReviewResponse{Success: true}, nil
}
