package service

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"sync"
	"testing"

	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
)

// =========================================================================
// HELPERS
// =========================================================================

func buildFileHeader(filename, content string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "text/csv")

	part, _ := writer.CreatePart(h)
	io.WriteString(part, content)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(1 << 20)
	return form.File["file"][0]
}

// validCSV3Rows returns a valid CSV with 3 rows: item 1 has 2 rows, item 2 has 1 row.
func validCSV3Rows() string {
	return strings.Join([]string{
		"external_id,item_id,movement_type,quantity,movement_time,note",
		"EXT-001,1,IN,10,2026-05-25T08:00:00Z,note1",
		"EXT-002,2,OUT,5,2026-05-25T09:00:00Z,note2",
		"EXT-003,1,IN,3,2026-05-25T10:00:00Z,note3",
	}, "\n")
}

// =========================================================================
// TEST CASE 1: Non-CSV file extension -> validateFile must reject immediately
// =========================================================================
func TestImportBatch_InvalidFileExtension_ShouldReturnValidationError(t *testing.T) {
	fileHeader := buildFileHeader("data.txt", "some,content\n1,2")

	svc := newMovementService(&mockMovementRepo{}, nil, nil, nil, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err == nil {
		t.Fatal("expected validation error for non-CSV file, got nil")
	}

	if result.Total != 0 || result.Success != 0 {
		t.Errorf("expected empty ImportBatchResult on validation failure, got: %+v", result)
	}
}

// =========================================================================
// TEST CASE 2: CSV with header only, no data rows -> zero totals, no error
// =========================================================================
func TestImportBatch_EmptyCSV_ShouldReturnZeroTotals(t *testing.T) {
	fileHeader := buildFileHeader("data.csv", "external_id,item_id,movement_type,quantity,movement_time,note\n")

	svc := newMovementService(&mockMovementRepo{}, nil, nil, nil, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error for empty CSV: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("expected Total=0, got %d", result.Total)
	}

	if result.Success != 0 || result.Rejected != 0 || result.Duplicate != 0 {
		t.Errorf("expected all counts=0, got %+v", result)
	}

	if len(result.FailedRows) != 0 {
		t.Errorf("expected no FailedRows, got %d", len(result.FailedRows))
	}
}

// =========================================================================
// TEST CASE 3: All rows valid, everything succeeds -> Success equals Total
// =========================================================================
func TestImportBatch_AllValidRows_ShouldReturnAllSuccess(t *testing.T) {
	fileHeader := buildFileHeader("data.csv", validCSV3Rows())

	mockRepo := &mockMovementRepo{
		getExistingExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]string, error) {
			return nil, nil
		},
		createBatchFn: func(ctx context.Context, movements []*entity.Movement) error {
			return nil
		},
	}
	mockItemSvc := &mockItemService{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 100}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected Total=3, got %d", result.Total)
	}

	if result.Success != 3 {
		t.Errorf("expected Success=3, got %d", result.Success)
	}

	if result.Rejected != 0 {
		t.Errorf("expected Rejected=0, got %d", result.Rejected)
	}

	if result.Duplicate != 0 {
		t.Errorf("expected Duplicate=0, got %d", result.Duplicate)
	}

	if len(result.FailedRows) != 0 {
		t.Errorf("expected no FailedRows, got %d", len(result.FailedRows))
	}
}

// =========================================================================
// TEST CASE 4: repo.Create returns ErrDuplicateMovement
//
//	-> ProcessOne maps it to StatusDuplicate
//	-> ImportBatch increments Duplicate counter and appends to FailedRows
//
// =========================================================================
func TestImportBatch_DuplicateExternalID_ShouldCountDuplicates(t *testing.T) {
	// 2 rows with the same item_id=1; the second row triggers a duplicate key error
	csv := strings.Join([]string{
		"external_id,item_id,movement_type,quantity,movement_time,note",
		"EXT-DUP,1,IN,10,2026-05-25T08:00:00Z,first",
		"EXT-DUP,1,IN,10,2026-05-25T08:01:00Z,second",
	}, "\n")

	fileHeader := buildFileHeader("data.csv", csv)

	mockRepo := &mockMovementRepo{
		getExistingExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]string, error) {
			return nil, nil // Intra-batch check will catch the duplicate
		},
		createBatchFn: func(ctx context.Context, movements []*entity.Movement) error {
			return nil
		},
	}
	mockItemSvc := &mockItemService{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 100}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected Total=2, got %d", result.Total)
	}

	if result.Success != 1 {
		t.Errorf("expected Success=1, got %d", result.Success)
	}

	if result.Duplicate != 1 {
		t.Errorf("expected Duplicate=1, got %d", result.Duplicate)
	}

	if result.Rejected != 0 {
		t.Errorf("expected Rejected=0, got %d", result.Rejected)
	}

	if len(result.FailedRows) != 1 {
		t.Fatalf("expected 1 FailedRow, got %d", len(result.FailedRows))
	}

	if result.FailedRows[0].Status != entity.StatusDuplicate {
		t.Errorf("expected FailedRows[0].Status=StatusDuplicate, got %v", result.FailedRows[0].Status)
	}
}

// =========================================================================
// TEST CASE 5: AdjustStock returns ErrInsufficientStock
//
//	-> not an AppError and not ErrDuplicateMovement
//	-> ProcessOne falls into the internal error branch -> StatusRejected
//
// =========================================================================
func TestImportBatch_InsufficientStock_ShouldCountRejected(t *testing.T) {
	csv := strings.Join([]string{
		"external_id,item_id,movement_type,quantity,movement_time,note",
		"EXT-REJ,1,OUT,9999,2026-05-25T08:00:00Z,over-withdrawal",
	}, "\n")

	fileHeader := buildFileHeader("data.csv", csv)

	mockRepo := &mockMovementRepo{
		getExistingExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]string, error) {
			return nil, nil
		},
		createBatchFn: func(ctx context.Context, movements []*entity.Movement) error {
			return nil
		},
	}
	mockItemSvc := &mockItemService{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 100}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected Total=1, got %d", result.Total)
	}

	if result.Success != 0 {
		t.Errorf("expected Success=0, got %d", result.Success)
	}

	if result.Rejected != 1 {
		t.Errorf("expected Rejected=1, got %d", result.Rejected)
	}

	if result.Duplicate != 0 {
		t.Errorf("expected Duplicate=0, got %d", result.Duplicate)
	}

	if len(result.FailedRows) != 1 {
		t.Fatalf("expected 1 FailedRow, got %d", len(result.FailedRows))
	}

	if result.FailedRows[0].ExternalID != "EXT-REJ" {
		t.Errorf("expected FailedRows[0].ExternalID='EXT-REJ', got '%s'", result.FailedRows[0].ExternalID)
	}

	if result.FailedRows[0].Status != entity.StatusRejected {
		t.Errorf("expected FailedRows[0].Status=StatusRejected, got %v", result.FailedRows[0].Status)
	}
}

// =========================================================================
// TEST CASE 6: Row fails Movement.Validate() due to quantity=0
//
//	-> ProcessOne returns StatusRejected + ErrBadRequest before entering the transaction
//	-> repo.Create must NOT be called
//
// =========================================================================
func TestImportBatch_InvalidQuantity_ShouldBeRejectedByValidation(t *testing.T) {
	csv := strings.Join([]string{
		"external_id,item_id,movement_type,quantity,movement_time,note",
		"EXT-ZERO,1,IN,0,2026-05-25T08:00:00Z,zero-qty",
	}, "\n")

	fileHeader := buildFileHeader("data.csv", csv)

	repoCalled := false
	mockRepo := &mockMovementRepo{
		getExistingExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]string, error) {
			return nil, nil
		},
		createBatchFn: func(ctx context.Context, movements []*entity.Movement) error {
			repoCalled = true
			return nil
		},
	}
	mockItemSvc := &mockItemService{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 100}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repoCalled {
		t.Error("expected repo.Create NOT to be called when validation fails")
	}

	if result.Success != 0 {
		t.Errorf("expected Success=0, got %d", result.Success)
	}

	if result.Rejected != 1 {
		t.Errorf("expected Rejected=1, got %d", result.Rejected)
	}

	if result.FailedRows[0].Status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", result.FailedRows[0].Status)
	}
}

// =========================================================================
// TEST CASE 7: Mixed results — 1 success, 1 duplicate, 1 rejected
//
//	-> each counter and FailedRows must be summarized correctly
//
// =========================================================================
func TestImportBatch_MixedResults_ShouldSummarizeCorrectly(t *testing.T) {
	csv := strings.Join([]string{
		"external_id,item_id,movement_type,quantity,movement_time,note",
		"EXT-OK,1,IN,10,2026-05-25T08:00:00Z,ok",
		"EXT-DUP,2,IN,5,2026-05-25T09:00:00Z,dup",
		"EXT-REJ,3,OUT,9999,2026-05-25T10:00:00Z,reject",
	}, "\n")

	fileHeader := buildFileHeader("data.csv", csv)

	mockRepo := &mockMovementRepo{
		getExistingExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]string, error) {
			for _, id := range externalIDs {
				if id == "EXT-DUP" {
					return []string{"EXT-DUP"}, nil
				}
			}
			return nil, nil
		},
		createBatchFn: func(ctx context.Context, movements []*entity.Movement) error {
			return nil
		},
	}
	mockItemSvc := &mockItemService{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			if id == 3 {
				return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 0}, nil
			}
			return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 100}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected Total=3, got %d", result.Total)
	}

	if result.Success != 1 {
		t.Errorf("expected Success=1, got %d", result.Success)
	}

	if result.Duplicate != 1 {
		t.Errorf("expected Duplicate=1, got %d", result.Duplicate)
	}

	if result.Rejected != 1 {
		t.Errorf("expected Rejected=1, got %d", result.Rejected)
	}

	if len(result.FailedRows) != 2 {
		t.Errorf("expected 2 FailedRows, got %d", len(result.FailedRows))
	}
}

// =========================================================================
// TEST CASE 8: Rows with the same item_id must be processed sequentially
//
//	-> order within the same worker group must be preserved
//
// =========================================================================
func TestImportBatch_SameItemRows_ShouldProcessSequentially(t *testing.T) {
	csv := strings.Join([]string{
		"external_id,item_id,movement_type,quantity,movement_time,note",
		"EXT-A,1,IN,10,2026-05-25T08:00:00Z,first",
		"EXT-B,1,IN,5,2026-05-25T09:00:00Z,second",
		"EXT-C,1,OUT,3,2026-05-25T10:00:00Z,third",
	}, "\n")

	fileHeader := buildFileHeader("data.csv", csv)

	var mu sync.Mutex
	var processOrder []string

	mockRepo := &mockMovementRepo{
		getExistingExternalIDsFn: func(ctx context.Context, externalIDs []string) ([]string, error) {
			return nil, nil
		},
		createBatchFn: func(ctx context.Context, movements []*entity.Movement) error {
			mu.Lock()
			for _, m := range movements {
				processOrder = append(processOrder, m.ExternalID)
			}
			mu.Unlock()
			return nil
		},
	}
	mockItemSvc := &mockItemService{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{SQLModel: core.SQLModel{ID: id}, CurrentStock: 100}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	result, err := svc.ImportBatch(context.Background(), fileHeader)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Success != 3 {
		t.Errorf("expected Success=3, got %d", result.Success)
	}

	expected := []string{"EXT-A", "EXT-B", "EXT-C"}
	if len(processOrder) != 3 {
		t.Fatalf("expected 3 processed rows, got %d", len(processOrder))
	}

	for i, id := range expected {
		if processOrder[i] != id {
			t.Errorf("expected processOrder[%d]=%s, got %s (sequential order violated)", i, id, processOrder[i])
		}
	}
}
