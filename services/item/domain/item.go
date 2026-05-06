package domain

import (
	"errors"
	"inventory-movement-processing/pkg/core"
	"strings"
)

var (
	ErrItemNameEmpty     = errors.New("item name cannot be empty")
	ErrItemSKUEmpty      = errors.New("item SKU cannot be empty")
	ErrInvalidStock      = errors.New("current stock cannot be negative")
	ErrInvalidThreshold  = errors.New("low stock threshold cannot be negative")
	ErrInvalidCategoryID = errors.New("category_id must be greater than 0")
	ErrCategoryNameEmpty = errors.New("category name cannot be empty")
)

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

	if i.CategoryID <=0 {
		return ErrInvalidCategoryID
	}
	return nil
}

func (c *Category) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return ErrCategoryNameEmpty
	}
	return nil
}