package postgres

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
)

func (r *movementRepository) GetMovementsByItemID(ctx context.Context, itemId int) ([]entity.Movement, error) {
	var movements []entity.Movement
	err := r.db.WithContext(ctx).Where("item_id = ?", itemId).Find(&movements).Error
	return movements, err
}
