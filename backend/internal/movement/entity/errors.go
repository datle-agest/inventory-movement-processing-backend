package entity

import "errors"

// Movement errors
var (
	ErrInvalidItemID     = errors.New("item_id must be greater than 0")
	ErrInvalidQuantity   = errors.New("quantity must be greater than 0")
	ErrInvalidType       = errors.New("movement_type must be IN, OUT or ADJUST")
	ErrExternalIDEmpty   = errors.New("external_id cannot be empty")
	ErrDuplicateMovement = errors.New("movement with this external_id already exists")
	ErrInsufficientStock = errors.New("insufficient stock balance")
)
