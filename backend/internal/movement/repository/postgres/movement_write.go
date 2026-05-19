package postgres

import (
	"context"
	"errors"
	inventoryEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *movementRepository) ProcessMovement(ctx context.Context, m *movementEntity.Movement) error {

	// begin transaction
	tx := r.db.WithContext(ctx).Begin()

	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if recover() != nil {
			tx.Rollback()
		}
	}()
	// lock inventory item row
	var item inventoryEntity.Item

	err := tx.
		Clauses(clause.Locking{
			Strength: "UPDATE",
		}).
		Where("id = ?", m.ItemID).
		First(&item).Error

	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return inventoryEntity.ErrItemNotFound
		}
		return err
	}

	// process stock movement
	switch m.Type {

	case movementEntity.MovementTypeOut:
		// check stock
		if item.CurrentStock < m.Quantity {
			tx.Rollback()
			return inventoryEntity.ErrInsufficientStock
		}
		item.CurrentStock -= m.Quantity

	case movementEntity.MovementTypeIn:
		item.CurrentStock += m.Quantity

	case movementEntity.MovementTypeAdjust:
		if m.Quantity < 0 && item.CurrentStock < -m.Quantity {
			tx.Rollback()
			return inventoryEntity.ErrInsufficientStock
		}
		item.CurrentStock += m.Quantity
	}

	// create inventory movement
	if err := tx.Create(m).Error; err != nil {
		tx.Rollback()
		// duplicate external_id
		if strings.Contains(
			strings.ToLower(err.Error()),
			"duplicate",
		) {

			return inventoryEntity.ErrDuplicateMovement
		}
		return err
	}

	// update inventory stock
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "chk_current_stock_non_negative") {
			return inventoryEntity.ErrInsufficientStock
		}
		return err
	}

	// commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
