package postgres

import (
	"context"
	itemEntity "inventory-movement-processing/internal/item/entity"
)

func (repo *repository) GetItemByIDs(ctx context.Context, ids []int32) ([]itemEntity.Item, error) {
	var items []itemEntity.Item

	if len(ids) == 0 {
		return items, nil
	}

	err := repo.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (repo *repository) ListLowStockItems(
	ctx context.Context,
) ([]*itemEntity.Item, error) {

	var results []*itemEntity.Item

	err := repo.db.WithContext(ctx).
		Where("low_stock_threshold > 0 AND current_stock < low_stock_threshold").
		Order("(low_stock_threshold - current_stock) DESC").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (repo *repository) ListItem(ctx context.Context) ([]itemEntity.Item, error) {
	var items []itemEntity.Item
	err := repo.db.WithContext(ctx).Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (repo *repository) GetItem(ctx context.Context, id int32) (*itemEntity.Item, error) {

	var item itemEntity.Item

	err := repo.db.WithContext(ctx).First(&item, id).Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}
