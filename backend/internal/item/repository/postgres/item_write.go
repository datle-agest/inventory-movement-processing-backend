package postgres

import (
	"context"
	"inventory-movement-processing/internal/item/entity"

	"gorm.io/gorm"
)

func (repo *repository) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {
	err := repo.db.WithContext(ctx).Create(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (repo *repository) DeleteItem(ctx context.Context, id int) error {
	result := repo.db.WithContext(ctx).Delete(&entity.Item{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
