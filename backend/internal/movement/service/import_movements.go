package service

import (
	"context"
	"encoding/csv"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
	"io"
	"mime/multipart"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *service) ImportBatch(ctx context.Context, file *multipart.FileHeader) (entity.ImportBatchResult, error) {
	// validate file
	if err := s.validateFile(file); err != nil {
		return entity.ImportBatchResult{}, err
	}

	// open file
	f, err := file.Open()
	if err != nil {
		return entity.ImportBatchResult{}, common.ErrInternal("cannot open file")
	}
	defer f.Close()

	// parse csv
	validRows, parseFailedRows, err := s.parseCSV(f)
	if err != nil {
		return entity.ImportBatchResult{}, err
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

// extractCSVHeaders - Extract csv tags from struct
func extractCSVHeaders(v any) []string {

	t := reflect.TypeOf(v)

	headers := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {

		tag := t.Field(i).Tag.Get("csv")

		// skip ignored/internal fields
		if tag == "" || tag == "-" || tag == "row_index" {
			continue
		}

		headers = append(headers, tag)
	}

	return headers
}

// validateCSVHeader - Validate CSV header format
func validateCSVHeader(actual []string) error {

	expectedHeaders := extractCSVHeaders(entity.CsvMovementRow{})

	if len(actual) != len(expectedHeaders) {
		return common.ErrBadRequest("invalid csv header")
	}

	for i, expected := range expectedHeaders {

		actualHeader := strings.TrimSpace(actual[i])

		if actualHeader != expected {
			return common.ErrBadRequest("invalid csv header format")
		}
	}

	return nil
}

// parseCSV - Parse CSV file
func (s *service) parseCSV(src io.Reader) ([]entity.CsvMovementRow, []entity.ProcessResult, error) {

	reader := csv.NewReader(src)

	// read header
	header, err := reader.Read()
	if err != nil {
		return nil, nil, common.ErrBadRequest("cannot read csv header")
	}

	if err := validateCSVHeader(header); err != nil {
		return nil, nil, err
	}

	var (
		rows       []entity.CsvMovementRow
		failedRows []entity.ProcessResult
		rowIndex   = 1
	)

	for {
		record, err := reader.Read()
		// valid header

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
				entity.ProcessResult{
					RowIndex:    rowIndex,
					ExternalID:  row["external_id"],
					Status:      entity.StatusRejected,
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
				entity.ProcessResult{
					RowIndex:    rowIndex,
					ExternalID:  row["external_id"],
					Status:      entity.StatusRejected,
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
				entity.ProcessResult{
					RowIndex:    rowIndex,
					ExternalID:  row["external_id"],
					Status:      entity.StatusRejected,
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
				entity.ProcessResult{
					RowIndex:   rowIndex,
					ExternalID: row["external_id"],
					Status:     entity.StatusRejected,
					ErrorReason: "invalid movement_time " +
						"(RFC3339 required)",
				},
			)
			continue
		}

		csvRow := entity.CsvMovementRow{
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
func (s *service) groupRowsByItem(rows []entity.CsvMovementRow) map[int32][]entity.CsvMovementRow {
	result := make(map[int32][]entity.CsvMovementRow)
	for _, row := range rows {
		result[row.ItemID] = append(result[row.ItemID], row)
	}
	return result
}

func (s *service) runImportWorkers(ctx context.Context, grouped map[int32][]entity.CsvMovementRow) chan entity.ProcessResult {
	totalRows := 0
	for _, rows := range grouped {
		totalRows += len(rows)
	}
	resultCh := make(chan entity.ProcessResult, totalRows)

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

				res := entity.ProcessResult{
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
	parseFailedRows []entity.ProcessResult,
	resultCh chan entity.ProcessResult,
) entity.ImportBatchResult {

	var (
		success   int
		rejected  int
		duplicate int
	)

	failedRows := append(
		[]entity.ProcessResult{},
		parseFailedRows...,
	)

	rejected = len(parseFailedRows)

	for r := range resultCh {

		switch r.Status {

		case entity.StatusAccepted:
			success++

		case entity.StatusRejected:
			rejected++
			failedRows = append(failedRows, r)

		case entity.StatusDuplicate:
			duplicate++
			failedRows = append(failedRows, r)
		}
	}

	return entity.ImportBatchResult{
		Total:      total,
		Success:    success,
		Rejected:   rejected,
		Duplicate:  duplicate,
		FailedRows: failedRows,
	}
}
