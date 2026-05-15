package service

import (
	"context"
	"encoding/csv"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
	"io"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ProcessStatus string

const (
	StatusAccepted  ProcessStatus = "accepted"
	StatusRejected  ProcessStatus = "rejected"
	StatusDuplicate ProcessStatus = "duplicate"
)

type csvMovementRow struct {
	RowIndex     int                 `csv:"row_index"`
	ExternalID   string              `csv:"external_id"`
	ItemID       int32               `csv:"item_id"`
	Type         entity.MovementType `csv:"movement_type"`
	Quantity     int32               `csv:"quantity"`
	MovementTime time.Time           `csv:"movement_time"` // RFC3339 format: 2026-05-15T08:00:00Z
	Note         string              `csv:"note"`
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

func (s *service) ImportBatch(ctx context.Context, file *multipart.FileHeader) (ImportBatchResult, error) {
	// validate file
	if err := s.validateFile(file); err != nil {
		return ImportBatchResult{}, err
	}

	// open file
	f, err := file.Open()
	if err != nil {
		return ImportBatchResult{}, common.ErrInternal("cannot open file")
	}
	defer f.Close()

	// parse csv
	validRows, parseFailedRows, err := s.parseCSV(f)
	if err != nil {
		return ImportBatchResult{}, err
	}

	// group by item_id
	groupedRows := s.groupRowsByItem(validRows)

	// run workers
	resultCh := s.runImportWorkers(ctx, groupedRows)

	totalRows := len(validRows) + len(parseFailedRows)

	// summarize
	result := s.summarizeResults(
		totalRows,
		parseFailedRows,
		resultCh,
	)

	return result, nil
}

// validateFile - Validate file
func (s *service) validateFile(file *multipart.FileHeader) error {
	if file == nil {
		return common.ErrBadRequest("file is required")
	}
	if file.Size == 0 {
		return common.ErrBadRequest("file is empty")
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return common.ErrBadRequest("file must be CSV format")
	}
	return nil
}

// parseCSV - Parse CSV file
func (s *service) parseCSV(src io.Reader) ([]csvMovementRow, []ProcessResult, error) {

	reader := csv.NewReader(src)

	// read header
	header, err := reader.Read()
	if err != nil {
		return nil, nil, common.ErrBadRequest("cannot read csv header")
	}

	var (
		rows       []csvMovementRow
		failedRows []ProcessResult
		rowIndex   = 1
	)

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		// hard csv error
		if err != nil {
			return nil, nil, common.ErrBadRequest("invalid csv format")
		}

		rowIndex++

		row := make(map[string]string)

		for i, h := range header {

			if i < len(record) {
				row[strings.TrimSpace(h)] = strings.TrimSpace(record[i])
			}
		}

		// validate item_id
		itemIDInt, err := strconv.Atoi(row["item_id"])
		if err != nil {
			failedRows = append(
				failedRows,
				ProcessResult{
					RowIndex:    rowIndex,
					ExternalID:  row["external_id"],
					Status:      StatusRejected,
					ErrorReason: "invalid item_id",
				},
			)
			continue
		}

		// validate quantity
		quantityInt, err := strconv.Atoi(row["quantity"])
		if err != nil {
			failedRows = append(
				failedRows,
				ProcessResult{
					RowIndex:    rowIndex,
					ExternalID:  row["external_id"],
					Status:      StatusRejected,
					ErrorReason: "invalid quantity",
				},
			)
			continue
		}

		movementType := entity.MovementType(row["movement_type"])
		// validate movement type
		if !movementType.IsValid() {
			failedRows = append(
				failedRows,
				ProcessResult{
					RowIndex:    rowIndex,
					ExternalID:  row["external_id"],
					Status:      StatusRejected,
					ErrorReason: "invalid movement_type",
				},
			)
			continue
		}

		// validate movement time
		movementTime, err := time.Parse(time.RFC3339, row["movement_time"])

		if err != nil {
			failedRows = append(
				failedRows,
				ProcessResult{
					RowIndex:   rowIndex,
					ExternalID: row["external_id"],
					Status:     StatusRejected,
					ErrorReason: "invalid movement_time " +
						"(RFC3339 required)",
				},
			)
			continue
		}

		csvRow := csvMovementRow{
			RowIndex:     rowIndex,
			ExternalID:   row["external_id"],
			ItemID:       int32(itemIDInt),
			Type:         movementType,
			Quantity:     int32(quantityInt),
			MovementTime: movementTime,
			Note:         row["note"],
		}
		rows = append(rows, csvRow)
	}
	// sort by business event time
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].MovementTime.Before(
			rows[j].MovementTime,
		)
	})

	return rows, failedRows, nil
}

// groupRowsByItem - Group rows by item_id for sequential processing per item
func (s *service) groupRowsByItem(rows []csvMovementRow) map[int32][]csvMovementRow {
	result := make(map[int32][]csvMovementRow)
	for _, row := range rows {
		result[row.ItemID] = append(result[row.ItemID], row)
	}
	return result
}

func (s *service) runImportWorkers(ctx context.Context, grouped map[int32][]csvMovementRow) chan ProcessResult {
	totalRows := 0
	for _, rows := range grouped {
		totalRows += len(rows)
	}
	resultCh := make(chan ProcessResult, totalRows)

	// Submit one job per item group
	// Movements of same item processed sequentially
	// Different items processed concurrently
	for _, itemRows := range grouped {

		rows := itemRows

		s.workerPool.Submit(func() {
			// sequential within same item
			for _, r := range rows {

				movement := &entity.Movement{
					ExternalID:   r.ExternalID,
					ItemID:       r.ItemID,
					Type:         r.Type,
					Quantity:     r.Quantity,
					MovementTime: r.MovementTime,
					Note:         &r.Note,
				}

				status, err := s.ProcessOne(
					ctx,
					movement,
				)

				res := ProcessResult{
					RowIndex:   r.RowIndex,
					ExternalID: r.ExternalID,
					Status:     status,
				}

				if err != nil {
					res.ErrorReason = err.Error()
				}

				resultCh <- res
			}
		})
	}

	go func() {
		s.workerPool.Wait()
		close(resultCh)
	}()

	return resultCh
}

// summarizeResults
func (s *service) summarizeResults(
	total int,
	parseFailedRows []ProcessResult,
	resultCh chan ProcessResult,
) ImportBatchResult {

	var (
		success   int
		rejected  int
		duplicate int
	)

	failedRows := append(
		[]ProcessResult{},
		parseFailedRows...,
	)

	rejected = len(parseFailedRows)

	for r := range resultCh {

		switch r.Status {

		case StatusAccepted:
			success++

		case StatusRejected:
			rejected++
			failedRows = append(failedRows, r)

		case StatusDuplicate:
			duplicate++
			failedRows = append(failedRows, r)
		}
	}

	return ImportBatchResult{
		Total:      total,
		Success:    success,
		Rejected:   rejected,
		Duplicate:  duplicate,
		FailedRows: failedRows,
	}
}
