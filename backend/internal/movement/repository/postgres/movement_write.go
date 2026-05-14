package postgres

import (
	"context"
	movementEntity "inventory-movement-processing/internal/movement/entity"
)

// Create - Tạo movement
func (r *movementRepository) Create(ctx context.Context, movement *movementEntity.Movement) error {
	if err := r.db.WithContext(ctx).Create(movement).Error; err != nil {
		return err
	}
	return nil
}

// CreateBatch - Tạo nhiều movements cùng lúc
func (r *movementRepository) CreateBatch(ctx context.Context, movements []*movementEntity.Movement) error {
	if err := r.db.WithContext(ctx).CreateInBatches(movements, 100).Error; err != nil {
		return err
	}
	return nil
}
