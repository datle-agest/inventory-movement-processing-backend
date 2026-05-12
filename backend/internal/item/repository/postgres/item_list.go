package postgres

import (
	"context"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
)

func (repo *repository) ListItem(ctx context.Context) ([]entity.Item, error) {
	var items []entity.Item
	err := repo.db.WithContext(ctx).Find(&items).Error

	if err != nil {
		return nil, common.ErrInternal("cannot list items")
	}

	return items, nil
}
