package postgres

import (
	"context"
	"fmt"
	"inventory-movement-processing/internal/item/entity"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"strings"
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

func (repo *repository) ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
	var items []entity.Item
	db := repo.db.WithContext(ctx).Model(&entity.Item{})

	if filter.Search != nil && *filter.Search != "" {
		searchText := "%" + *filter.Search + "%"
		db = db.Where("name ILIKE ? OR sku ILIKE ?", searchText, searchText)
	}

	if filter.LowStock != nil && *filter.LowStock {
		db = db.Where("low_stock_threshold > 0 AND current_stock < low_stock_threshold")
	}

	if filter.OutOfStock != nil && *filter.OutOfStock {
		db = db.Where("current_stock = 0")
	}

	if filter.MinQty != nil {
		db = db.Where("current_stock >= ?", *filter.MinQty)
	}

	if filter.MaxQty != nil {
		db = db.Where("current_stock <= ?", *filter.MaxQty)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}
	paging.Total = int(total)

	allowedSortFields := map[string]string{
		"id":                  "id",
		"name":                "name",
		"sku":                 "sku",
		"current_stock":       "current_stock",
		"low_stock_threshold": "low_stock_threshold",
		"created_at":          "created_at",
	}

	sortColumn, exists := allowedSortFields[strings.ToLower(filter.SortBy)]
	if !exists {
		sortColumn = "id"
	}

	sortDirection := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		sortDirection = "ASC"
	}

	db = db.Order(fmt.Sprintf("%s %s", sortColumn, sortDirection))

	offset := (paging.Page - 1) * paging.Limit
	err := db.Offset(offset).Limit(paging.Limit).Find(&items).Error
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
