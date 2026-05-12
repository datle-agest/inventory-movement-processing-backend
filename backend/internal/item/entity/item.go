package entity

import (
	"inventory-movement-processing/pkg/core"
	"strings"
)

type Item struct {
	core.SQLModel
	Name              string `json:"name"                gorm:"column:name;type:varchar(255);not null"`
	SKU               string `json:"sku"                 gorm:"column:sku;type:varchar(100);uniqueIndex;not null"`
	CurrentStock      int32  `json:"current_stock"       gorm:"column:current_stock;not null;default:0;check:chk_current_stock_non_negative,current_stock >= 0"`
	LowStockThreshold int32  `json:"low_stock_threshold" gorm:"column:low_stock_threshold;not null;default:0;check:chk_low_stock_threshold_non_negative,low_stock_threshold >= 0"`
}

func (Item) TableName() string {
	return "inventory_items"
}

func (i *Item) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return ErrItemNameEmpty
	}

	if strings.TrimSpace(i.SKU) == "" {
		return ErrItemSKUEmpty
	}

	if i.CurrentStock < 0 {
		return ErrInvalidStock
	}

	if i.LowStockThreshold < 0 {
		return ErrInvalidThreshold
	}

	return nil
}
