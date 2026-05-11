package entity

import "errors"

var (
	ErrInvalidItemID   = errors.New("item_id must be greater than 0")
	ErrInvalidQuantity = errors.New("quantity must be greater than 0")
	ErrInvalidType     = errors.New("movement_type must be IN, OUT or ADJUST")
)
