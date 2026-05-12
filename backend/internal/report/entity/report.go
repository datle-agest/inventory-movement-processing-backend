package entity

import (
	"inventory-movement-processing/pkg/core"
	"time"
)

type ReportItem struct {
	ItemID   int32  `json:"item_id"`
	Name     string `json:"name,omitempty"`
	Quantity int32  `json:"quantity,omitempty"`
}

type Report struct {
	core.SQLModel

	ReportDate            time.Time `json:"report_date"`
	TotalInCount          int32     `json:"total_in_count"`
	TotalOutCount         int32     `json:"total_out_count"`
	TotalAdjustCount      int32     `json:"total_adjust_count"`
	TotalQuantityReceived int32     `json:"total_qty_received"`
	TotalQuantityIssued   int32     `json:"total_qty_issued"`

	Top5ActiveItem []ReportItem `gorm:"type:jsonb" json:"top5_active_items"`
	LowStockItem   []ReportItem `gorm:"type:jsonb" json:"low_stock_items"`
}

func (Report) TableName() string {
	return "daily_inventory_reports"
}

func (r *Report) Validate() error {
	if r.ReportDate.IsZero() {
		return ErrReportDate
	}

	if r.TotalInCount < 0 || r.TotalOutCount < 0 || r.TotalAdjustCount < 0 || r.TotalQuantityReceived < 0 || r.TotalQuantityIssued < 0 {
		return ErrNegativeReportVal
	}
	return nil
}
