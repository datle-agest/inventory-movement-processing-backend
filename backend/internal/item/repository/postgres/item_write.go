package postgres

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/components/gormc"

	"gorm.io/gorm"
)

func (repo *repository) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {
	err := repo.db.WithContext(ctx).Create(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, entity.ErrItemDuplicated
		}
		return nil, err
	}
	return &item, nil
}

// UpdateStock - Cập nhật stock (IN/OUT/ADJUST)
func (repo *repository) UpdateStock(ctx context.Context, itemID int32, newStock int32) error {
	db := gormc.GetDB(ctx, repo.db)
	result := db.WithContext(ctx).
		Model(&entity.Item{}).
		Where("id = ?", itemID).
		Update("current_stock", newStock)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
