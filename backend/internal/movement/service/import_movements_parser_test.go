package service

import (
	"bytes"
	"inventory-movement-processing/internal/movement/entity"
	"mime/multipart"
	"strings"
	"testing"
	"time"
)

// =========================================================================
// TEST SUITE: validateFile
// =========================================================================

// =========================================================================
// TEST CASE 1: File is Nil -> Should return BadRequest
// =========================================================================
func TestValidateFile_NilFile_ShouldReturnBadRequest(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	err := svc.validateFile(nil)

	if err == nil {
		t.Fatalf("expected error when file is nil, got nil")
	}

	expectedMsg := "file is required"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to contain '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 2: File Size is Zero -> Should return BadRequest
// =========================================================================
func TestValidateFile_EmptySize_ShouldReturnBadRequest(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	fileHeader := &multipart.FileHeader{
		Filename: "test.csv",
		Size:     0,
	}

	err := svc.validateFile(fileHeader)

	if err == nil {
		t.Fatalf("expected error when file size is 0, got nil")
	}

	expectedMsg := "file is empty"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to contain '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 3: File Extension is Not CSV -> Should return BadRequest
// =========================================================================
func TestValidateFile_InvalidExtension_ShouldReturnBadRequest(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	invalidFiles := []string{"data.txt", "report.xlsx", "backup.zip", "csv_hidden.png"}

	for _, filename := range invalidFiles {
		fileHeader := &multipart.FileHeader{
			Filename: filename,
			Size:     1024,
		}

		err := svc.validateFile(fileHeader)

		if err == nil {
			t.Fatalf("expected error for invalid extension '%s', got nil", filename)
		}

		expectedMsg := "file must be CSV format"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error message to contain '%s', got: '%v'", expectedMsg, err)
		}
	}
}

// =========================================================================
// TEST CASE 4: File Valid -> Should return Nil Error
// =========================================================================
func TestValidateFile_ValidCSV_ShouldReturnNil(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	validFiles := []string{"data.csv", "REPORT.CSV", "inventory_2026.Csv"}

	for _, filename := range validFiles {
		fileHeader := &multipart.FileHeader{
			Filename: filename,
			Size:     512,
		}

		err := svc.validateFile(fileHeader)

		if err != nil {
			t.Errorf("unexpected error for valid filename '%s': %v", filename, err)
		}
	}
}

// =========================================================================
// TEST SUITE: parseCSV
// =========================================================================

// =========================================================================
// TEST CASE 5: Cannot Read CSV Header / Corrupted Structure -> Should return BadRequest
// =========================================================================
func TestParseCSV_CorruptedOrEmptyReader_ShouldReturnBadRequest(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	var buf bytes.Buffer

	validRows, failedRows, err := svc.parseCSV(&buf)

	if validRows != nil || failedRows != nil {
		t.Errorf("expected outputs to be nil upon initialization failure")
	}

	if err == nil {
		t.Fatalf("expected error when reading empty content, got nil")
	}

	expectedMsg := "cannot read csv header"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to contain '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 6: Invalid CSV Header Schema Format -> Should return BadRequest
// =========================================================================
func TestParseCSV_InvalidHeaderFormat_ShouldReturnBadRequest(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	invalidCSVContent := "wrong_id,sku,qty,time_stamp\n" +
		"1,EXT-001,100,2026-05-25T00:00:00Z\n"

	buf := bytes.NewBufferString(invalidCSVContent)

	_, _, err := svc.parseCSV(buf)

	if err == nil {
		t.Fatalf("expected validation error due to invalid csv header format, got nil")
	}
}

// =========================================================================
// TEST CASE 7: Row-Level Failures -> Should Populate failedRows Slice
// =========================================================================
func TestParseCSV_RowValidationFailures_ShouldCaptureInFailedRows(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	headers := extractCSVHeaders(entity.CsvMovementRow{})
	headerLine := strings.Join(headers, ",")

	// Row 2: invalid item_id (string instead of int)
	// Row 3: invalid quantity (float instead of int)
	// Row 4: invalid movement_type (UNKNOWN type)
	// Row 5: invalid movement_time (non-RFC3339 layout)
	csvContent := headerLine + "\n" +
		"EXT-101,abc,IN,50,2026-05-25T09:00:00Z,Valid item token\n" +
		"EXT-102,200,OUT,12.5,2026-05-25T09:15:00Z,Fractional item quantity\n" +
		"EXT-103,300,INVALID_TYPE,10,2026-05-25T09:30:00Z,Invalid enum value\n" +
		"EXT-104,400,IN,5,25-05-2026 09:45:00,Invalid timestamp format\n"

	buf := bytes.NewBufferString(csvContent)

	validRows, failedRows, err := svc.parseCSV(buf)

	if err != nil {
		t.Fatalf("unexpected parsing lifecycle abort error: %v", err)
	}

	if len(validRows) != 0 {
		t.Errorf("expected 0 valid rows to match database model, got %d", len(validRows))
	}

	if len(failedRows) != 4 {
		t.Fatalf("expected exactly 4 failed records captured, got %d", len(failedRows))
	}

	if failedRows[0].RowIndex != 2 || failedRows[0].ErrorReason != "invalid item_id" {
		t.Errorf("unexpected parsing error evaluation context at row 2: %v", failedRows[0])
	}

	if failedRows[1].RowIndex != 3 || failedRows[1].ErrorReason != "invalid quantity" {
		t.Errorf("unexpected parsing error evaluation context at row 3: %v", failedRows[1])
	}

	if failedRows[2].RowIndex != 4 || failedRows[2].ErrorReason != "invalid movement_type" {
		t.Errorf("unexpected parsing error evaluation context at row 4: %v", failedRows[2])
	}

	if failedRows[3].RowIndex != 5 || !strings.Contains(failedRows[3].ErrorReason, "invalid movement_time") {
		t.Errorf("unexpected parsing error evaluation context at row 5: %v", failedRows[3])
	}
}

// =========================================================================
// TEST CASE 8: Success Flow -> Should Parse and Sort by Chronological Order
// =========================================================================
func TestParseCSV_Success_ShouldParseCorrectly(t *testing.T) {
	mockLog := &mockLogger{}
	svc := newMovementService(nil, nil, nil, nil, mockLog).(*service)

	headers := extractCSVHeaders(entity.CsvMovementRow{})
	headerLine := strings.Join(headers, ",")

	csvContent := headerLine + "\n" +
		"EXT-AAA,501,IN,100,2026-05-25T10:00:00Z,Late item block\n" +
		"EXT-BBB,502,OUT,20,2026-05-25T08:00:00Z,Early item block\n" +
		"EXT-CCC,503,IN,50,2026-05-25T09:00:00Z,Mid item block\n"

	buf := bytes.NewBufferString(csvContent)

	validRows, failedRows, err := svc.parseCSV(buf)

	if err != nil {
		t.Fatalf("unexpected validation routing process error: %v", err)
	}

	if len(failedRows) != 0 {
		t.Fatalf("expected 0 row parsing failures, got %d", len(failedRows))
	}

	if len(validRows) != 3 {
		t.Fatalf("expected exactly 3 successfully parsed entries, got %d", len(validRows))
	}

	expectedTime0, _ := time.Parse(time.RFC3339, "2026-05-25T10:00:00Z")
	expectedTime1, _ := time.Parse(time.RFC3339, "2026-05-25T08:00:00Z")
	expectedTime2, _ := time.Parse(time.RFC3339, "2026-05-25T09:00:00Z")

	if !validRows[0].MovementTime.Equal(expectedTime0) || validRows[0].ExternalID != "EXT-AAA" {
		t.Errorf("expected index 0 to contain EXT-AAA, got: %s", validRows[0].ExternalID)
	}

	if !validRows[1].MovementTime.Equal(expectedTime1) || validRows[1].ExternalID != "EXT-BBB" {
		t.Errorf("expected index 1 to contain EXT-BBB, got: %s", validRows[1].ExternalID)
	}

	if !validRows[2].MovementTime.Equal(expectedTime2) || validRows[2].ExternalID != "EXT-CCC" {
		t.Errorf("expected index 2 to contain EXT-CCC, got: %s", validRows[2].ExternalID)
	}

	if validRows[1].ItemID != 502 || validRows[1].Quantity != 20 || validRows[1].Type != "OUT" {
		t.Errorf("data conversions failed to map values onto target struct cleanly")
	}
}