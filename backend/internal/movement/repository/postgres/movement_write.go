package postgres

import (
	"context"
	"errors"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/components/gormc"

	"gorm.io/gorm"
)

// Create - Tạo movement
func (r *movementRepository) Create(ctx context.Context, movement *movementEntity.Movement) error {
	// Lấy DB từ transaction context
	db := gormc.GetDB(ctx, r.db)

	if err := db.WithContext(ctx).
		Create(movement).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return itemEntity.ErrDuplicateMovement
		}
		return err
	}

	return nil
}

// CreateBatch - Tạo nhiều movements cùng lúc
func (r *movementRepository) CreateBatch(ctx context.Context, movements []*movementEntity.Movement) error {
	if err := r.db.WithContext(ctx).CreateInBatches(movements, 100).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return itemEntity.ErrDuplicateMovement
		}
		return err
	}
	return nil
}
