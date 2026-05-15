package entity

import (
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
	"time"
)

type DailyItemSummary struct {
	core.SQLModel

	SummaryDate time.Time `gorm:"column:summary_date;type:date;not null;uniqueIndex:idx_item_summary_date"`
	ItemID      int32     `gorm:"column:item_id;not null;uniqueIndex:idx_item_summary_date"`

	Item *itemEntity.Item `gorm:"foreignKey:ItemID;references:ID"`

	TotalIn     int32 `gorm:"column:total_in;not null;default:0"`
	TotalOut    int32 `gorm:"column:total_out;not null;default:0"`
	TotalAdjust int32 `gorm:"column:total_adjust;not null;default:0"`
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
