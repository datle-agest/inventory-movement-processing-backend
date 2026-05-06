package domain

import "inventory-movement-processing/pkg/core"

// TODO: define category ????
type Item struct {
	core.SQLModel
	Name              string    `json:"name"`
	SKU               string    `json:"sku"`
	CurrentStock      int32     `json:"current_stock"`
	LowStockThreshold int32     `json:"low_stock_threshold"`
	CategoryID        int32     `json:"category_id"`
	Category          *Category `json:"category,omitempty"`
}

func (Item) TableName() string {
	return "inventory_items"
}

type Category struct {
	core.SQLModel
	Name  string `json:"name"`
	Items []Item `json:"items,omitempty"`
}

func (Category) TableName() string {
	return "inventory_categories"
}
