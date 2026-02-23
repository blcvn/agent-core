package handlers

import (
	"net/http"

	knowledgepb "github.com/blcvn/ba-shared-libs/proto/knowledge"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	knowledgeClient knowledgepb.KnowledgeServiceClient
}

func NewDocumentHandler(client knowledgepb.KnowledgeServiceClient) *DocumentHandler {
	return &DocumentHandler{knowledgeClient: client}
}

// CreatePRD handles POST /documents/prd
func (h *DocumentHandler) CreatePRD(c *gin.Context) {
	var req knowledgepb.CreatePRDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.knowledgeClient.CreatePRD(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GenerateDocument handles POST /documents/generate
func (h *DocumentHandler) GenerateDocument(c *gin.Context) {
	var req knowledgepb.GenerateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.knowledgeClient.GenerateDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetDocument handles GET /documents/:id
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	req := &knowledgepb.GetDocumentRequest{DocumentId: id}
	resp, err := h.knowledgeClient.GetDocument(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateDocument handles PUT /documents/:id
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	var req knowledgepb.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DocumentId = id // ensure ID matches URL

	resp, err := h.knowledgeClient.UpdateDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ApproveDocument handles POST /documents/:id/approve
func (h *DocumentHandler) ApproveDocument(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	req := &knowledgepb.ApproveDocumentRequest{DocumentId: id}
	resp, err := h.knowledgeClient.ApproveDocument(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReviewDocument handles POST /documents/:id/review
func (h *DocumentHandler) ReviewDocument(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	var req knowledgepb.ReviewDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DocumentId = id

	resp, err := h.knowledgeClient.ReviewDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetReviewStatus handles GET /reviews/:id
func (h *DocumentHandler) GetReviewStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	req := &knowledgepb.GetReviewStatusRequest{ReviewId: id}
	resp, err := h.knowledgeClient.GetReviewStatus(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RegenerateDocument handles POST /documents/:id/regenerate
func (h *DocumentHandler) RegenerateDocument(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	var req knowledgepb.RegenerateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DocumentId = id

	resp, err := h.knowledgeClient.RegenerateDocument(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
