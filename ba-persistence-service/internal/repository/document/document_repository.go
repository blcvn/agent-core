package document

import (
	"context"

	v32 "github.com/blcvn/backend/services/ba-persistence-service/internal/domain/v3.2"
	"gorm.io/gorm"
)

// DocumentRepository implements v32.DocumentRepository
type DocumentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create inserts a new document
func (r *DocumentRepository) Create(ctx context.Context, doc *v32.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// Get retrieves a document by ID
func (r *DocumentRepository) Get(ctx context.Context, id string) (*v32.Document, error) {
	var doc v32.Document
	if err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// Get retrieves a document by ID
func (r *DocumentRepository) GetByParentId(ctx context.Context, parentId string, tier v32.RequirementTier) (*v32.Document, error) {
	var doc v32.Document
	if err := r.db.WithContext(ctx).First(&doc, "parent_document_id = ? and tier = ?", parentId, tier).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// Update updates an existing document
func (r *DocumentRepository) Update(ctx context.Context, doc *v32.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// Delete removes a document by ID
func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&v32.Document{}, "id = ?", id).Error
}

// ListByProject retrieves all documents for a project
func (r *DocumentRepository) ListByProject(ctx context.Context, projectID string) ([]*v32.Document, error) {
	var docs []*v32.Document
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// GetByParent retrieves documents by parent ID
func (r *DocumentRepository) GetByParent(ctx context.Context, parentID string) ([]*v32.Document, error) {
	var docs []*v32.Document
	if err := r.db.WithContext(ctx).Where("parent_document_id = ?", parentID).Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// Verify interface compliance
var _ v32.DocumentRepository = (*DocumentRepository)(nil)
