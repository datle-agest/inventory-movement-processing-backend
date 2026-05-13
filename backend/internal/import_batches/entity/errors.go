package entity

import "errors"

var (
	ErrFileNameEmpty      = errors.New("file_name cannot be empty")
	ErrInvalidTotalRows   = errors.New("total_rows must be greater than 0")
	ErrBatchNotFound      = errors.New("batch not found")
	ErrBatchAlreadyExists = errors.New("batch already exists")
)
