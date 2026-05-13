package entity

import (
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
	"time"
)

type DailyItemSummary struct {
	core.SQLModel
	SummaryDate time.Time        `json:"summary_date" gorm:"column:summary_date;type:date;index;not null"`
	ItemID      int32            `json:"item_id"     gorm:"column:item_id;not null;index"`
	Item        *itemEntity.Item `json:"item,omitempty" gorm:"foreignKey:ItemID;references:ID"`
	TotalIn     int32            `json:"total_in"    gorm:"column:total_in;not null;default:0"`
	TotalOut    int32            `json:"total_out"   gorm:"column:total_out;not null;default:0"`
	TotalAdjust int32            `json:"total_adjust" gorm:"column:total_adjust;not null;default:0"`
}

func (DailyItemSummary) TableName() string {
	return "daily_item_summary"
}

func (dis *DailyItemSummary) Validate() error {
	if dis.ItemID <= 0 {
		return entity.ErrInvalidItemID
	}
	if dis.SummaryDate.IsZero() {
		return ErrInvalidDate
	}
	return nil
}

// Helper: Tính net stock change
func (dis *DailyItemSummary) NetChange() int32 {
	return dis.TotalIn - dis.TotalOut + dis.TotalAdjust
}
