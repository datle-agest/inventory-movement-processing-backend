package entity

import "time"

type ProcessStatus string

const (
	StatusAccepted  ProcessStatus = "accepted"
	StatusRejected  ProcessStatus = "rejected"
	StatusDuplicate ProcessStatus = "duplicate"
)

type CsvMovementRow struct {
	RowIndex     int          `csv:"row_index"`
	ExternalID   string       `csv:"external_id"`
	ItemID       int32        `csv:"item_id"`
	Type         MovementType `csv:"movement_type"`
	Quantity     int32        `csv:"quantity"`
	MovementTime time.Time    `csv:"movement_time"` // RFC3339 format: 2026-05-15T08:00:00Z
	Note         string       `csv:"note"`
}

type ProcessResult struct {
	RowIndex    int           `json:"row_index"`
	ExternalID  string        `json:"external_id"`
	Status      ProcessStatus `json:"status"`
	ErrorReason string        `json:"error_reason,omitempty"`
}

type ImportBatchResult struct {
	Total      int             `json:"total"`
	Success    int             `json:"success"`
	Rejected   int             `json:"rejected"`
	Duplicate  int             `json:"duplicate"`
	FailedRows []ProcessResult `json:"failed_rows"`
}
