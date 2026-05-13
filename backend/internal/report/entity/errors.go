package entity

import "errors"

var (
	ErrReportDate        = errors.New("report date is required")
	ErrNegativeReportVal = errors.New("report counts or quantities cannot be negative")
	ErrReportNotFound    = errors.New("report not found")
)

// DailyItemSummary errors
var (
	ErrInvalidDate        = errors.New("summary_date cannot be empty")
	ErrSummaryNotFound    = errors.New("daily summary not found")
	ErrInvalidSummaryDate = errors.New("summary date must be valid date")
)
