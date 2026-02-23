package postgres

import (
	"context"

	v32 "github.com/blcvn/backend/services/pkg/domain/v3.2"
	"gorm.io/gorm"
)

type ApprovalRepository struct {
	db *gorm.DB
}

func NewApprovalRepository(db *gorm.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

func (r *ApprovalRepository) Create(ctx context.Context, approval *v32.Approval) error {
	return r.db.WithContext(ctx).Create(approval).Error
}

func (r *ApprovalRepository) GetByID(ctx context.Context, id string) (*v32.Approval, error) {
	var approval v32.Approval
	if err := r.db.WithContext(ctx).First(&approval, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *ApprovalRepository) GetByDocumentID(ctx context.Context, documentID string) ([]*v32.Approval, error) {
	var approvals []*v32.Approval
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}
