package entity

import "errors"

// Movement errors
var (
	ErrInvalidItemID     = errors.New("item_id must be greater than 0")
	ErrInvalidQuantity   = errors.New("quantity must be greater than 0")
	ErrInvalidType       = errors.New("movement_type must be IN, OUT or ADJUST")
	ErrExternalIDEmpty   = errors.New("external_id cannot be empty")
	ErrDuplicateMovement = errors.New("movement with this external_id already exists")
)

// ImportBatch errors
var (
	ErrFileNameEmpty      = errors.New("file_name cannot be empty")
	ErrInvalidTotalRows   = errors.New("total_rows must be greater than 0")
	ErrBatchNotFound      = errors.New("batch not found")
	ErrBatchAlreadyExists = errors.New("batch already exists")
)
