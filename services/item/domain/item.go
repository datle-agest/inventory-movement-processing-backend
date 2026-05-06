package domain

import "inventory-movement-processing/pkg/core"

// TODO: define category ????
type Item struct {
	core.SQLModel
	Name     string `json:"name"`
	Category string `json:"category"`
	Quantity int32  `json:"quantity"`
}

func (Item) TableName() string {
	return "inventory_items"
}
