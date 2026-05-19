package postgres

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/components/gormc"
	"inventory-movement-processing/pkg/core"

	"gorm.io/gorm/clause"
)

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

func (repo *repository) ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
	var items []entity.Item

	db := repo.db.WithContext(ctx).Model(&entity.Item{}).Scopes(
		buildFilterScopes(filter)...,
	)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	paging.Total = int(total)

	offset := (paging.Page - 1) * paging.Limit
	err := db.Scopes(withSorting(filter.SortBy, filter.SortOrder)).
		Offset(offset).
		Limit(paging.Limit).
		Find(&items).Error

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

func (repo *repository) GetItemForUpdate(ctx context.Context, id int32) (*itemEntity.Item, error) {

	var item itemEntity.Item

	db := gormc.GetDB(ctx, repo.db)

	err := db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&item, id).Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}
