package entity

import (
	"inventory-movement-processing/pkg/core"
	"strings"
)

type Item struct {
	core.SQLModel
	Name              string    `json:"name"`
	SKU               string    `json:"sku"`
	CurrentStock      int32     `json:"current_stock"`
	LowStockThreshold int32     `json:"low_stock_threshold"`
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
