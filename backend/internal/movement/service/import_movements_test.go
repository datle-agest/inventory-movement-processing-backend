package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"testing"

	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
)

// Helper to create a multipart.FileHeader in memory for tests
func createTestMultipartFileHeader(t *testing.T, filename string, content string) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatalf("failed to create mock request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}

	fileHeaders := req.MultipartForm.File["file"]
	if len(fileHeaders) == 0 {
		t.Fatal("no file headers parsed")
	}

	return fileHeaders[0]
}

func TestImportBatch_ValidationErrors(t *testing.T) {
	svc := newMovementService(nil, nil)

	// Case 1: nil file
	_, err := svc.ImportBatch(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil file, got nil")
	}
	var appErr *common.AppError
	if errors.As(err, &appErr) && appErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected bad request, got status: %d", appErr.StatusCode)
	}

	// Case 2: empty file (size 0)
	emptyHeader := &multipart.FileHeader{
		Filename: "empty.csv",
		Size:     0,
	}
	_, err = svc.ImportBatch(context.Background(), emptyHeader)
	if err == nil {
		t.Error("expected error for empty file, got nil")
	}

	// Case 3: non-CSV file extension
	nonCsvHeader := createTestMultipartFileHeader(t, "data.txt", "some plain text data")
	_, err = svc.ImportBatch(context.Background(), nonCsvHeader)
	if err == nil {
		t.Error("expected error for non-CSV file extension, got nil")
	}
}

func TestImportBatch_InvalidCSVFormat(t *testing.T) {
	svc := newMovementService(nil, nil)

	// CSV with unclosed quotes or corrupt structure
	invalidCsvHeader := createTestMultipartFileHeader(t, "corrupt.csv", `external_id,item_id
"unclosed_quote,12`)

	_, err := svc.ImportBatch(context.Background(), invalidCsvHeader)
	if err == nil {
		t.Fatal("expected error parsing invalid CSV, got nil")
	}
}

func TestImportBatch_ParseFailedRows(t *testing.T) {
	csvContent := `external_id,item_id,movement_type,quantity,movement_time,note
EXT-001,invalid_item_id,IN,10,2026-05-15T08:00:00Z,Invalid Item ID
EXT-002,100,IN,invalid_quantity,2026-05-15T08:00:00Z,Invalid Quantity
EXT-003,100,INVALID_TYPE,10,2026-05-15T08:00:00Z,Invalid Type
EXT-004,100,IN,10,invalid_time,Invalid Time`

	fileHeader := createTestMultipartFileHeader(t, "data.csv", csvContent)

	svc := newMovementService(nil, &mockWorkerPool{})
	result, err := svc.ImportBatch(context.Background(), fileHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 4 {
		t.Errorf("expected 4 total rows, got %d", result.Total)
	}
	if result.Success != 0 {
		t.Errorf("expected 0 success rows, got %d", result.Success)
	}
	if result.Rejected != 4 {
		t.Errorf("expected 4 rejected rows, got %d", result.Rejected)
	}
	if len(result.FailedRows) != 4 {
		t.Fatalf("expected 4 failed rows, got %d", len(result.FailedRows))
	}

	// Verify specific failures
	expectedReasons := []string{
		"invalid item_id",
		"invalid quantity",
		"invalid movement_type",
		"invalid movement_time",
	}

	for i, failedRow := range result.FailedRows {
		if !bytes.Contains([]byte(failedRow.ErrorReason), []byte(expectedReasons[i])) {
			t.Errorf("row %d: expected error reason containing '%s', got '%s'", i+1, expectedReasons[i], failedRow.ErrorReason)
		}
	}
}

func TestImportBatch_SuccessAndBusinessErrors(t *testing.T) {
	// A CSV with a mix of:
	// - EXT-001: Success
	// - EXT-002: Insufficient stock (Rejected in business rule)
	// - EXT-003: Duplicate movement (Duplicate in business rule)
	// - EXT-004: Success
	csvContent := `external_id,item_id,movement_type,quantity,movement_time,note
EXT-001,1,IN,50,2026-05-15T08:00:00Z,Valid In
EXT-002,1,OUT,100,2026-05-15T09:00:00Z,Insufficient stock Out
EXT-003,2,IN,10,2026-05-15T10:00:00Z,Duplicate In
EXT-004,2,ADJUST,-5,2026-05-15T11:00:00Z,Valid Adjustment`

	fileHeader := createTestMultipartFileHeader(t, "data.csv", csvContent)

	mr := &mockMovementRepo{
		processMovementFn: func(ctx context.Context, m *movementEntity.Movement) error {
			switch m.ExternalID {
			case "EXT-001":
				return nil
			case "EXT-002":
				return itemEntity.ErrInsufficientStock
			case "EXT-003":
				return itemEntity.ErrDuplicateMovement
			case "EXT-004":
				return nil
			}
			return errors.New("unexpected external id")
		},
	}

	svc := newMovementService(mr, &mockWorkerPool{})
	result, err := svc.ImportBatch(context.Background(), fileHeader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 4 {
		t.Errorf("expected 4 total rows, got %d", result.Total)
	}
	if result.Success != 2 {
		t.Errorf("expected 2 success rows, got %d", result.Success)
	}
	if result.Rejected != 1 { // EXT-002
		t.Errorf("expected 1 rejected row, got %d", result.Rejected)
	}
	if result.Duplicate != 1 { // EXT-003
		t.Errorf("expected 1 duplicate row, got %d", result.Duplicate)
	}

	// Check failed rows details
	if len(result.FailedRows) != 2 {
		t.Fatalf("expected 2 failed rows, got %d", len(result.FailedRows))
	}

	// Failed row index sorting or ordering
	for _, failedRow := range result.FailedRows {
		if failedRow.ExternalID == "EXT-002" {
			if failedRow.Status != StatusRejected {
				t.Errorf("expected EXT-002 status rejected, got %s", failedRow.Status)
			}
			if failedRow.ErrorReason != "insufficient stock" {
				t.Errorf("expected insufficient stock, got '%s'", failedRow.ErrorReason)
			}
		} else if failedRow.ExternalID == "EXT-003" {
			if failedRow.Status != StatusDuplicate {
				t.Errorf("expected EXT-003 status duplicate, got %s", failedRow.Status)
			}
			if failedRow.ErrorReason != "duplicate external_id" {
				t.Errorf("expected duplicate external_id, got '%s'", failedRow.ErrorReason)
			}
		} else {
			t.Errorf("unexpected failed row: %+v", failedRow)
		}
	}
}
