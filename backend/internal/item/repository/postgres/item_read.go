package postgres

import (
	"context"
	"inventory-movement-processing/internal/item/entity"
)

func (repo *repository) GetItem(ctx context.Context, id int) (*entity.Item, error) {

	var item entity.Item

	err := repo.db.WithContext(ctx).First(&item, id).Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}
