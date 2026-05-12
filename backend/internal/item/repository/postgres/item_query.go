package postgres

import (
	"context"
	itemEntity "inventory-movement-processing/internal/item/entity"
)

func (r *repository) GetItemByIDs(ctx context.Context, ids []int32) ([]itemEntity.Item, error) {
	var items []itemEntity.Item

	if len(ids) == 0 {
		return items, nil
	}

	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) GetLowStockItems(ctx context.Context) ([]itemEntity.Item, error) {
	var items []itemEntity.Item

	err := r.db.WithContext(ctx).
		Where("current_stock <= low_stock_threshold").
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}
