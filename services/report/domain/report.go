package domain

import (
	"errors"
	"inventory-movement-processing/pkg/core"
	"time"
)

var (
	ErrReportDate        = errors.New("report date is required")
	ErrNegativeReportVal = errors.New("report counts or quantities cannot be negative")
)

type Report struct {
	core.SQLModel
	ReportDate            time.Time `json:"report_date"`
	TotalInCount          int       `json:"total_in_count"`
	TotalOutCount         int       `json:"total_out_count"`
	TotalAdjustCount      int       `json:"total_adjust_count"`
	TotalQuantityReceived int       `json:"total_qty_received"`
	TotalQuantityIssued   int       `json:"total_qty_issued"`
	Top5ActiveItem        string    `json:"top5_active_items"`
	LowStockItem          string    `json:"low_stock_items"`
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
