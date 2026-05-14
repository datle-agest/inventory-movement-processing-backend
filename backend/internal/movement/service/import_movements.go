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
	ExternalID string `csv:"external_id"`
	ItemID     int32  `csv:"item_id"`
	Type       string `csv:"movement_type"`
	Quantity   int32  `csv:"quantity"`
	Note       string `csv:"note"`
}

type ProcessResult struct {
	RowIndex    int           `json:"row_index"`
	ExternalID  string        `json:"external_id"`
	Status      ProcessStatus `json:"status"`
	ErrorReason string        `json:"error_reason,omitempty"`
}

func (s *service) ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus {
	// 1. Validate row
	// 2. Check external_id exists
	// 3. Check stock availability (cho OUT)
	s.mu.Lock()
	// 4. Create movement

	// 5. Update item stock (ATOMIC - database sẽ handle concurrency)

	// 6. COMMIT

	s.mu.Unlock()

	// validate
	if err := m.Validate(); err != nil {
		return StatusRejected
	}

	return StatusAccepted
}

func (s *service) ImportBatch(ctx context.Context, file *multipart.FileHeader) (map[string]interface{}, error) {
	// 1 validate

	// 2. Parse CSV
	// 3. Setup Worker Pool
	// 4. Tạo workers
	// 5. Gửi rows vào input channel (goroutine riêng)
	// 6. Nhận results
	// 7. Chờ tất cả workers xong
	return nil, nil
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
func (s *service) parseCSV(src io.Reader) ([]csvMovementRow, error) {
	reader := csv.NewReader(src)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, common.ErrBadRequest("cannot read csv header")
	}

	var rows []csvMovementRow

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, common.ErrBadRequest("invalid csv format")
		}

		row := make(map[string]string)

		for i, h := range header {
			if i < len(record) {
				row[strings.TrimSpace(h)] = strings.TrimSpace(record[i])
			}
		}

		itemIDInt, err := strconv.Atoi(row["item_id"])
		if err != nil {
			return nil, common.ErrBadRequest("invalid item_id")
		}

		quantityInt, err := strconv.Atoi(row["quantity"])
		if err != nil {
			return nil, common.ErrBadRequest("invalid quantity")
		}

		csvRow := csvMovementRow{
			ExternalID: row["external_id"],
			ItemID:     int32(itemIDInt),
			Type:       row["movement_type"],
			Quantity:   int32(quantityInt),
			Note:       row["note"],
		}

		rows = append(rows, csvRow)
	}

	return rows, nil
}
