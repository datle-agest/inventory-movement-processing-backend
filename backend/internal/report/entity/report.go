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

	ReportDate            time.Time `json:"report_date"         gorm:"column:report_date;type:timestamptz;not null;uniqueIndex"`
	TotalInCount          int32     `json:"total_in_count"      gorm:"column:total_in_count;not null;default:0"`
	TotalOutCount         int32     `json:"total_out_count"     gorm:"column:total_out_count;not null;default:0"`
	TotalAdjustCount      int32     `json:"total_adjust_count"  gorm:"column:total_adjust_count;not null;default:0"`
	TotalQuantityReceived int32     `json:"total_qty_received"  gorm:"column:total_qty_received;not null;default:0"`
	TotalQuantityIssued   int32     `json:"total_qty_issued"    gorm:"column:total_qty_issued;not null;default:0"`

	Top5ActiveItem []ReportItem `json:"top5_active_items"   gorm:"column:top5_active_items;type:jsonb;serializer:json"`
	LowStockItem   []ReportItem `json:"low_stock_items"     gorm:"column:low_stock_items;type:jsonb;serializer:json"`
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
