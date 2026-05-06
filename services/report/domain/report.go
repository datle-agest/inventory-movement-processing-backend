package domain

import (
	"inventory-movement-processing/pkg/core"
)

type Report struct {
	core.SQLModel
	TotalInCount          int    `json:"total_in_count"`
	TotalOutCount         int    `json:"total_out_count"`
	TotalAdjustCount      int    `json:"total_adjust_count"`
	TotalQuantityReceived int    `json:"total_qty_received"`
	TotalQuantityIssued   int    `json:"total_qty_issued"`
	Top5ActiveItem        string `json:"top5_active_items"`
	LowStockItem          string `json:"low_stock_items"`
}

func (Report) TableName() string {
	return "daily_inventory_reports"
}
