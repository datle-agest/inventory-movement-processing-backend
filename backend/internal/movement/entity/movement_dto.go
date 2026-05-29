package entity

import (
	"time"
)

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
	RowIndex    int           `json:"row_index" example:"18"`
	ExternalID  string        `json:"external_id" example:"TXN-120017"`
	Status      ProcessStatus `json:"status" example:"rejected" enums:"accepted,rejected,duplicate"`
	ErrorReason string        `json:"error_reason,omitempty" example:"item not found"`
}

type ImportBatchResult struct {
	Total      int             `json:"total" example:"1000"`
	Success    int             `json:"success" example:"999"`
	Rejected   int             `json:"rejected" example:"1"`
	Duplicate  int             `json:"duplicate" example:"0"`
	FailedRows []ProcessResult `json:"failed_rows"`
}

// NewMovementFromCSV maps a CsvMovementRow to a Movement entity.
func NewMovementFromCSV(r CsvMovementRow) *Movement {
	return &Movement{
		ExternalID:   r.ExternalID,
		ItemID:       r.ItemID,
		Type:         r.Type,
		Quantity:     r.Quantity,
		MovementTime: r.MovementTime,
		Note:         &r.Note,
	}
}

// NewAcceptedResult creates a ProcessResult with StatusAccepted.
func NewAcceptedResult(r CsvMovementRow) ProcessResult {
	return ProcessResult{
		RowIndex:   r.RowIndex,
		ExternalID: r.ExternalID,
		Status:     StatusAccepted,
	}
}

// NewRejectedResult creates a ProcessResult with the given status and error reason.
func NewRejectedResult(r CsvMovementRow, status ProcessStatus, reason string) ProcessResult {
	return ProcessResult{
		RowIndex:    r.RowIndex,
		ExternalID:  r.ExternalID,
		Status:      status,
		ErrorReason: reason,
	}
}
