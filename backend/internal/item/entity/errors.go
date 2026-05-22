package entity

import "errors"

var (
	ErrItemNameEmpty     = errors.New("item name cannot be empty")
	ErrItemSKUEmpty      = errors.New("item SKU cannot be empty")
	ErrInvalidStock      = errors.New("current stock cannot be negative")
	ErrInvalidThreshold  = errors.New("low stock threshold cannot be negative")
	ErrInvalidCategoryID = errors.New("category_id must be greater than 0")
	ErrCategoryNameEmpty = errors.New("category name cannot be empty")
	ErrItemNotFound      = errors.New("item not found")
	ErrInsufficientStock = errors.New("insufficient stock balance")
	ErrDuplicateMovement = errors.New("movement with this external_id already exists")
	ErrItemDuplicated    = errors.New("item with this SKU already exists")
)
