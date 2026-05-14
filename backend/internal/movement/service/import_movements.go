package service

import (
	"context"
	"encoding/csv"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
)

type ProcessStatus string

const (
	StatusAccepted  ProcessStatus = "accepted"
	StatusRejected  ProcessStatus = "rejected"
	StatusDuplicate ProcessStatus = "duplicate"
)

type csvMovementRow struct {
	ExternalID string              `csv:"external_id"`
	ItemID     int32               `csv:"item_id"`
	Type       entity.MovementType `csv:"movement_type"`
	Quantity   int32               `csv:"quantity"`
	Note       string              `csv:"note"`
}

type ProcessResult struct {
	RowIndex    int           `json:"row_index"`
	ExternalID  string        `json:"external_id"`
	Status      ProcessStatus `json:"status"`
	ErrorReason string        `json:"error_reason,omitempty"`
}

func (s *service) ImportBatch(ctx context.Context, file *multipart.FileHeader) (map[string]interface{}, error) {
	// validate file
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// open file
	f, err := file.Open()
	if err != nil {
		return nil, common.ErrInternal("cannot open file")
	}
	defer f.Close()

	// parse csv
	validRows, parseFailedRows, err := s.parseCSV(f)
	if err != nil {
		return nil, err
	}

	// run worker
	resultCh := s.runImportWorkers(ctx, validRows)

	totalRows := len(validRows) + len(parseFailedRows)

	result := s.summarizeResults(totalRows, parseFailedRows, resultCh)

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
		itemIDInt, err := strconv.Atoi(
			row["item_id"],
		)

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
		quantityInt, err := strconv.Atoi(
			row["quantity"],
		)

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

		movementType := entity.MovementType(
			row["movement_type"],
		)

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

		csvRow := csvMovementRow{
			ExternalID: row["external_id"],
			ItemID:     int32(itemIDInt),
			Type:       movementType,
			Quantity:   int32(quantityInt),
			Note:       row["note"],
		}

		rows = append(rows, csvRow)
	}

	return rows, failedRows, nil
}

func (s *service) runImportWorkers(ctx context.Context, rows []csvMovementRow) chan ProcessResult {
	resultCh := make(chan ProcessResult, len(rows))

	for idx, row := range rows {
		i, r := idx, row // Tránh lỗi closure
		s.workerPool.Submit(func() {
			movement := &entity.Movement{
				ExternalID: r.ExternalID,
				ItemID:     r.ItemID,
				Type:       r.Type,
				Quantity:   r.Quantity,
				Note:       &r.Note,
			}
			status, err := s.ProcessOne(ctx, movement)

			res := ProcessResult{RowIndex: i + 1, ExternalID: r.ExternalID, Status: status}
			if err != nil {
				res.ErrorReason = err.Error()
			}
			resultCh <- res
		})
	}

	go func() {
		s.workerPool.Wait()
		close(resultCh)
	}()

	return resultCh
}

func (s *service) summarizeResults(total int,
	parseFailedRows []ProcessResult,
	resultCh chan ProcessResult,
) map[string]interface{} {

	var success, rejected, duplicate int
	// Khởi tạo mảng failedRows chứa sẵn các lỗi từ lúc parseCSV
	failedRows := append([]ProcessResult{}, parseFailedRows...)

	// Cập nhật số lượng rejected ban đầu bằng số lượng lỗi parse
	rejected = len(parseFailedRows)

	// Đọc tiếp kết quả từ các worker (những dòng hợp lệ đã chạy xong)
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

	return map[string]interface{}{
		"total":       total,
		"success":     success,
		"rejected":    rejected,
		"duplicate":   duplicate,
		"failed_rows": failedRows,
	}
}
