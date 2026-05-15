package entity

import (
	itemEntity "inventory-movement-processing/internal/item/entity"
)

type TopActiveItemsResult struct {
	TopItems      []*DailyItemSummary `json:"top_items"`
	LowStockItems []*itemEntity.Item  `json:"low_stock_items,omitempty"`
}
