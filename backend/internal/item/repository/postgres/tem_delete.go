package postgres

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (repo *repository) DeleteItem(ctx context.Context, id int) error {
	result := repo.db.WithContext(ctx).Delete(&entity.Item{}, id)
	if result.Error != nil {
		return common.ErrInternal("cannot delete item")
	}

	if result.RowsAffected == 0 {
		return common.ErrNotFound("item not found")
	}

	return nil
}
