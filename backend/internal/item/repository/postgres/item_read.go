package postgres

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"

	"gorm.io/gorm"
)

func (repo *repository) GetItem(ctx context.Context, id int) (*entity.Item, error) {
	var item entity.Item

	err := repo.db.WithContext(ctx).First(&item, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound("item not found")
		}
		return nil, common.ErrInternal("cannot get item")
	}

	return &item, nil
}
