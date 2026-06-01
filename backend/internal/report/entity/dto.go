package entity

import (
	"inventory-movement-processing/internal/item/entity"
)

type TopActiveItemsResult struct {
	TopItems      []*DailyItemSummary `json:"top_items"`
	LowStockItems []*entity.Item  `json:"low_stock_items,omitempty"`
}
