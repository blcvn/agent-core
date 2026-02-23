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

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
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
		Id:         doc.ID,
		Title:      doc.ModuleName, // Mapping ModuleName to Title for now
		Content:    doc.Content,
		Author:     "", // TODO: Add Author to v32.Document?
		ProjectId:  doc.ProjectID,
		ParentId:   doc.ParentDocumentID,
		RootId:     doc.RootDocumentID,
		ModuleName: doc.ModuleName,
		Tier:       doc.Tier.String(),
		Status:     string(doc.Status),
		Version:    int32(doc.Version),
		CreatedAt:  timestamppb.New(doc.CreatedAt),
		UpdatedAt:  timestamppb.New(doc.UpdatedAt),
	}, nil
}

// ListDocuments retrieves documents with optional filters
func (s *PersistenceServer) ListDocuments(ctx context.Context, req *persistencepb.ListDocumentsRequest) (*persistencepb.ListDocumentsResponse, error) {
	var docs []*v32.Document
	var err error

	// If parent_id and tier are specified, use GetByParentId for exact match
	if req.ParentId != "" && req.Tier != "" {
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

		doc, err := s.docRepo.GetByParentId(ctx, req.ParentId, tier)
		if err != nil {
			return nil, fmt.Errorf("failed to get document by parent and tier: %w", err)
		}
		docs = []*v32.Document{doc}
	} else if req.ParentId != "" {
		// List all children of a parent
		docs, err = s.docRepo.GetByParent(ctx, req.ParentId)
		if err != nil {
			return nil, fmt.Errorf("failed to list documents by parent: %w", err)
		}
	} else if req.ProjectId != "" {
		// List by project
		docs, err = s.docRepo.ListByProject(ctx, req.ProjectId)
		if err != nil {
			return nil, fmt.Errorf("failed to list documents by project: %w", err)
		}
	} else {
		return &persistencepb.ListDocumentsResponse{}, nil // No filter = empty
	}

	var pbDocs []*persistencepb.GetDocumentResponse
	for _, doc := range docs {
		pbDocs = append(pbDocs, &persistencepb.GetDocumentResponse{
			Id:         doc.ID,
			Title:      doc.ModuleName,
			Content:    doc.Content,
			ProjectId:  doc.ProjectID,
			ParentId:   doc.ParentDocumentID,
			RootId:     doc.RootDocumentID,
			ModuleName: doc.ModuleName,
			Tier:       doc.Tier.String(),
			Status:     string(doc.Status),
			Version:    int32(doc.Version),
			CreatedAt:  timestamppb.New(doc.CreatedAt),
			UpdatedAt:  timestamppb.New(doc.UpdatedAt),
		})
	}

	return &persistencepb.ListDocumentsResponse{Documents: pbDocs}, nil
}

// UpdateDocument updates a document's content and/or status
func (s *PersistenceServer) UpdateDocument(ctx context.Context, req *persistencepb.UpdateDocumentRequest) (*persistencepb.UpdateDocumentResponse, error) {
	// Fetch existing document
	doc, err := s.docRepo.Get(ctx, req.DocumentId)
	if err != nil {
		return &persistencepb.UpdateDocumentResponse{
			Success: false,
			Error:   fmt.Sprintf("document not found: %v", err),
		}, nil
	}

	// Apply updates
	if req.Content != "" {
		doc.Content = req.Content
	}
	if req.Status != "" {
		doc.Status = v32.DocumentStatus(req.Status)
	}
	doc.UpdatedAt = time.Now()
	doc.Version++

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return &persistencepb.UpdateDocumentResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &persistencepb.UpdateDocumentResponse{Success: true}, nil
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
// Review Implementation
func (s *PersistenceServer) CreateReview(ctx context.Context, req *persistencepb.CreateReviewRequest) (*persistencepb.CreateReviewResponse, error) {
	// Fetch document to get Tier
	doc, err := s.docRepo.Get(ctx, req.DocumentId)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	review := &v32.Review{
		ID:         fmt.Sprintf("rev_%d", time.Now().UnixNano()),
		DocumentID: req.DocumentId,
		Tier:       doc.Tier,
		Comment:    req.Comments, // Ignoring req.Reviewer as v32.Review lacks it
		ActionType: v32.ActionPending,
		CreatedAt:  time.Now(),
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return &persistencepb.CreateReviewResponse{
		Success: true,
	}, nil
}

func (s *PersistenceServer) ListReviews(ctx context.Context, req *persistencepb.ListReviewsRequest) (*persistencepb.ListReviewsResponse, error) {
	reviews, err := s.reviewRepo.GetByDocumentID(ctx, req.DocumentId)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}

	var pbReviews []*persistencepb.Review
	for _, r := range reviews {
		pbReviews = append(pbReviews, &persistencepb.Review{
			Id:         r.ID,
			DocumentId: r.DocumentID,
			Reviewer:   "", // Not stored
			Comments:   r.Comment,
			Status:     string(r.ActionType), // Mapping ActionType to Status/ActionType
			ActionType: string(r.ActionType),
			CreatedAt:  timestamppb.New(r.CreatedAt),
		})
	}

	return &persistencepb.ListReviewsResponse{
		Reviews: pbReviews,
	}, nil
}
