package postgres

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

func (repo *repository) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {
	err := repo.db.WithContext(ctx).Create(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
