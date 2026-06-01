package postgres

import (
	"fmt"
	"inventory-movement-processing/internal/item/entity"
	"strings"

	"gorm.io/gorm"
)

func buildFilterScopes(filter *entity.ItemFilter) []func(*gorm.DB) *gorm.DB {
	var scopes []func(*gorm.DB) *gorm.DB

	if filter == nil {
		return scopes
	}

	if filter.Name != nil && *filter.Name != "" {
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where("name ILIKE ?", "%"+*filter.Name+"%")
		})
	}

	if filter.SKU != nil && *filter.SKU != "" {
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where("sku = ?", *filter.SKU)
		})
	}

	if filter.LowStock != nil && *filter.LowStock {
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where("low_stock_threshold > 0 AND current_stock < low_stock_threshold")
		})
	}

	return scopes
}

func withSorting(sortBy, sortOrder string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		allowedSortFields := map[string]string{
			"id":                  "id",
			"name":                "name",
			"sku":                 "sku",
			"current_stock":       "current_stock",
			"low_stock_threshold": "low_stock_threshold",
			"created_at":          "created_at",
		}

		sortColumn, exists := allowedSortFields[strings.ToLower(sortBy)]
		if !exists {
			sortColumn = "id"
		}

		sortDirection := "DESC"
		if strings.ToLower(sortOrder) == "asc" {
			sortDirection = "ASC"
		}

		return db.Order(fmt.Sprintf("%s %s", sortColumn, sortDirection))
	}
}
