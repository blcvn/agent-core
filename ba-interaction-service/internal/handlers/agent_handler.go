package handlers

import (
	"context"
	"net/http"
	"time"

	orchestratorpb "github.com/blcvn/ba-shared-libs/proto/orchestrator"
	"github.com/gin-gonic/gin"
)

// AgentHandler handles HTTP requests and delegates to Orchestrator Service
type AgentHandler struct {
	orchestratorClient orchestratorpb.OrchestratorServiceClient
}

func NewAgentHandler(client orchestratorpb.OrchestratorServiceClient) *AgentHandler {
	return &AgentHandler{orchestratorClient: client}
}

// ExecuteTask godoc
// @Summary Submit a new agent task
// @Description Submit a task for the BA agent system to execute
// @Accept json
// @Produce json
// @Param request body ExecuteTaskRequest true "Task request"
// @Success 200 {object} ExecuteTaskResponse
// @Router /api/v1/tasks [post]
func (h *AgentHandler) ExecuteTask(c *gin.Context) {
	var req ExecuteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.orchestratorClient.ExecuteTask(ctx, &orchestratorpb.ExecuteTaskRequest{
		SessionId: req.SessionID,
		UserId:    req.UserID,
		Content:   req.Content,
		Metadata:  req.Metadata,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ExecuteTaskResponse{
		TaskID: resp.TaskId,
		Status: resp.Status,
	})
}

// GetTaskStatus godoc
// @Summary Get task status
// @Description Retrieve the current status and result of a task
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} TaskStatusResponse
// @Router /api/v1/tasks/{task_id} [get]
func (h *AgentHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.orchestratorClient.GetTaskStatus(ctx, &orchestratorpb.GetTaskStatusRequest{
		TaskId: taskID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, TaskStatusResponse{
		TaskID: resp.TaskId,
		Status: resp.Status,
		Result: resp.Result,
	})
}

// SubmitInput godoc
// @Summary Submit input for a waiting task
// @Description Provide user input for a task that is waiting for input
// @Accept json
// @Produce json
// @Param task_id path string true "Task ID"
// @Param request body SubmitInputRequest true "Input data"
// @Success 200 {object} SubmitInputResponse
// @Router /api/v1/tasks/{task_id}/input [post]
func (h *AgentHandler) SubmitInput(c *gin.Context) {
	taskID := c.Param("task_id")

	var req SubmitInputRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.orchestratorClient.SubmitInput(ctx, &orchestratorpb.SubmitInputRequest{
		TaskId:    taskID,
		InputData: req.InputData,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SubmitInputResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}

// CancelTask godoc
// @Summary Cancel a running task
// @Description Cancel an in-progress task
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {object} CancelTaskResponse
// @Router /api/v1/tasks/{task_id}/cancel [post]
func (h *AgentHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.orchestratorClient.CancelTask(ctx, &orchestratorpb.CancelTaskRequest{
		TaskId: taskID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CancelTaskResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}

// ── Request/Response Types ──

type ExecuteTaskRequest struct {
	SessionID string            `json:"session_id" binding:"required"`
	UserID    string            `json:"user_id" binding:"required"`
	Content   string            `json:"content" binding:"required"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ExecuteTaskResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type TaskStatusResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
	Result string `json:"result,omitempty"`
}

type SubmitInputRequest struct {
	InputData string `json:"input_data" binding:"required"`
}

type SubmitInputResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CancelTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
