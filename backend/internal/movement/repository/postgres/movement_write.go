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
	db := gormc.GetDB(ctx, r.db)
	if err := db.WithContext(ctx).CreateInBatches(movements, 100).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return itemEntity.ErrDuplicateMovement
		}
		return err
	}
	return nil
}

// SetLockTimeout configures the lock timeout for the current transaction
func (r *movementRepository) SetLockTimeout(ctx context.Context, timeout string) error {
	db := gormc.GetDB(ctx, r.db)
	// Using Sprintf because SET command parameters cannot be parameterized in pgx natively
	// The timeout variable should be trusted internal constant like "3s"
	query := "SET LOCAL lock_timeout = '" + timeout + "';"
	return db.WithContext(ctx).Exec(query).Error
}
