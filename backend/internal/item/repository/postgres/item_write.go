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

// UpdateStock - Cập nhật stock (IN/OUT/ADJUST)
func (repo *repository) UpdateStock(ctx context.Context, itemID int32, quantity int32) error {
	if err := repo.db.WithContext(ctx).Model(&entity.Item{}).
		Where("id = ?", itemID).
		Update("current_stock", gorm.Expr("current_stock + ?", quantity)).Error; err != nil {
		return err
	}
	return nil
}
