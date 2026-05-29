package service

import (
	"encoding/csv"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/movement/entity"
	"io"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// validateFile - Kiểm tra định dạng và tính toàn vẹn của file upload
func (s *service) validateFile(file *multipart.FileHeader) error {
	if file == nil {
		return common.NewBadRequestError(common.CodeFileRequired, "file is required")
	}
	if file.Size == 0 {
		return common.NewBadRequestError(common.CodeFileEmpty, "file is empty")
	}
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return common.NewBadRequestError(common.CodeFileMustBeCSV, "file must be CSV format")
	}
	return nil
}

// extractCSVHeaders - Đọc các tag `csv` được định nghĩa trong struct của Entity
func extractCSVHeaders(v any) []string {
	t := reflect.TypeOf(v)
	headers := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("csv")
		if tag == "" || tag == "-" || tag == "row_index" {
			continue
		}
		headers = append(headers, tag)
	}
	return headers
}

// validateCSVHeader - Đối chiếu header của file CSV upload xem có đúng định dạng chuẩn không
func validateCSVHeader(actual []string) error {
	expectedHeaders := extractCSVHeaders(entity.CsvMovementRow{})
	if len(actual) != len(expectedHeaders) {
		return common.NewBadRequestError(common.CodeInvalidCSVHeader, "invalid csv header")
	}
	for i, expected := range expectedHeaders {
		actualHeader := strings.TrimSpace(actual[i])
		if actualHeader != expected {
			return common.NewBadRequestError(common.CodeInvalidCSVHeader, "invalid csv header format")
		}
	}
	return nil
}

// parseCSV - Đọc file CSV, validate kiểu dữ liệu từng cột và parse thành Struct
func (s *service) parseCSV(src io.Reader) ([]entity.CsvMovementRow, []entity.ProcessResult, error) {
	reader := csv.NewReader(src)
	// Đọc và validate hàng Header đầu tiên
	header, err := reader.Read()
	if err != nil {
		return nil, nil, common.NewBadRequestError(common.CodeInvalidCSVFormat, "cannot read csv header")
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
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, common.NewBadRequestError(common.CodeInvalidCSVFormat, "invalid csv format")
		}
		rowIndex++
		row := make(map[string]string)
		for i, h := range header {
			if i < len(record) {
				row[strings.TrimSpace(h)] = strings.TrimSpace(record[i])
			}
		}
		// Validate cột item_id
		itemIDInt, err := strconv.Atoi(row["item_id"])
		if err != nil {
			s.logger.Warnf("[Service][parseCSV] row %d: invalid item_id '%s'", rowIndex, row["item_id"])
			failedRows = append(failedRows, entity.ProcessResult{
				RowIndex:    rowIndex,
				ExternalID:  row["external_id"],
				Status:      entity.StatusRejected,
				ErrorReason: "invalid item_id",
			})
			continue
		}
		// Validate cột quantity
		quantityInt, err := strconv.Atoi(row["quantity"])
		if err != nil {
			s.logger.Warnf("[Service][parseCSV] row %d: invalid quantity '%s'", rowIndex, row["quantity"])
			failedRows = append(failedRows, entity.ProcessResult{
				RowIndex:    rowIndex,
				ExternalID:  row["external_id"],
				Status:      entity.StatusRejected,
				ErrorReason: "invalid quantity",
			})
			continue
		}
		// Validate cột movement_type
		movementType := entity.MovementType(row["movement_type"])
		if !movementType.IsValid() {
			s.logger.Warnf("[Service][parseCSV] row %d: invalid movement_type '%s'", rowIndex, row["movement_type"])
			failedRows = append(failedRows, entity.ProcessResult{
				RowIndex:    rowIndex,
				ExternalID:  row["external_id"],
				Status:      entity.StatusRejected,
				ErrorReason: "invalid movement_type",
			})
			continue
		}
		// Validate cột movement_time
		movementTime, err := time.Parse(time.RFC3339, row["movement_time"])
		if err != nil {
			s.logger.Warnf("[Service][parseCSV] row %d: invalid movement_type '%s'", rowIndex, row["movement_type"])
			failedRows = append(failedRows, entity.ProcessResult{
				RowIndex:    rowIndex,
				ExternalID:  row["external_id"],
				Status:      entity.StatusRejected,
				ErrorReason: "invalid movement_time (RFC3339 required)",
			})
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
	return rows, failedRows, nil
}
