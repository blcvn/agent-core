package review

import (
	"context"
	"fmt"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
	"gorm.io/gorm"
)

// ReviewRepository implements v32.ReviewRepository
type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, review *v32.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *ReviewRepository) GetByDocumentID(ctx context.Context, documentID string) ([]*v32.Review, error) {
	var reviews []*v32.Review
	if err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Order("created_at desc").Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("failed to get reviews for document %s: %w", documentID, err)
	}
	return reviews, nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id string) (*v32.Review, error) {
	var review v32.Review
	if err := r.db.WithContext(ctx).First(&review, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get review %s: %w", id, err)
	}
	return &review, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *v32.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}
