package entity

import "errors"

var (
	ErrReportDate        = errors.New("report date is required")
	ErrNegativeReportVal = errors.New("report counts or quantities cannot be negative")
)
