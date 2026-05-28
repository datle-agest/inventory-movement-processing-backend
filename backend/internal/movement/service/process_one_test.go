package service

import (
	"context"
	"errors"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	"strings"
	"testing"
	"time"
)

// validMovement returns a fully valid Movement to use as a base in tests.
func validMovement() *entity.Movement {
	note := "test note"
	return &entity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       1,
		Type:         entity.MovementTypeIn,
		Quantity:     10,
		MovementTime: time.Now(),
		Note:         &note,
	}
}

// =========================================================================
// TEST CASE 1: Empty ExternalID -> Validate fails -> StatusRejected + BadRequest
// =========================================================================
func TestProcessOne_EmptyExternalID_ShouldReturnRejected(t *testing.T) {
	m := validMovement()
	m.ExternalID = ""

	svc := newMovementService(&mockMovementRepo{}, nil, nil, nil, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), entity.ErrExternalIDEmpty.Error()) {
		t.Errorf("expected error to contain '%v', got '%v'", entity.ErrExternalIDEmpty, err)
	}
}

// =========================================================================
// TEST CASE 2: ItemID <= 0 -> Validate fails -> StatusRejected + BadRequest
// =========================================================================
func TestProcessOne_InvalidItemID_ShouldReturnRejected(t *testing.T) {
	m := validMovement()
	m.ItemID = 0

	svc := newMovementService(&mockMovementRepo{}, nil, nil, nil, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), entity.ErrInvalidItemID.Error()) {
		t.Errorf("expected error to contain '%v', got '%v'", entity.ErrInvalidItemID, err)
	}
}

// =========================================================================
// TEST CASE 3: Quantity = 0 for MovementTypeIn -> Validate fails -> StatusRejected
// =========================================================================
func TestProcessOne_ZeroQuantity_ShouldReturnRejected(t *testing.T) {
	m := validMovement()
	m.Type = entity.MovementTypeIn
	m.Quantity = 0

	svc := newMovementService(&mockMovementRepo{}, nil, nil, nil, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), entity.ErrInvalidQuantity.Error()) {
		t.Errorf("expected error to contain '%v', got '%v'", entity.ErrInvalidQuantity, err)
	}
}

// =========================================================================
// TEST CASE 4: Zero MovementTime -> Validate fails -> StatusRejected + BadRequest
// =========================================================================
func TestProcessOne_ZeroMovementTime_ShouldReturnRejected(t *testing.T) {
	m := validMovement()
	m.MovementTime = time.Time{}

	svc := newMovementService(&mockMovementRepo{}, nil, nil, nil, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), entity.ErrInvalidMovementTime.Error()) {
		t.Errorf("expected error to contain '%v', got '%v'", entity.ErrInvalidMovementTime, err)
	}
}

// =========================================================================
// TEST CASE 5: AdjustStock returns ErrInsufficientStock
//
//	-> not an AppError, not ErrDuplicateMovement
//	-> falls into internal error branch -> StatusRejected + masked error
//
// =========================================================================
func TestProcessOne_InsufficientStock_ShouldReturnRejectedWithMaskedError(t *testing.T) {
	m := validMovement()
	m.Type = entity.MovementTypeOut

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			return itemEntity.ErrInsufficientStock
		},
	}

	svc := newMovementService(&mockMovementRepo{}, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "cannot process movement, please try again"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected masked internal error '%s', got '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 6: repo.Create returns ErrDuplicateMovement
//
//	-> errors.Is match -> StatusDuplicate + ErrConflict
//
// =========================================================================
func TestProcessOne_DuplicateMovement_ShouldReturnStatusDuplicate(t *testing.T) {
	m := validMovement()

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			return nil
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			return itemEntity.ErrDuplicateMovement
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusDuplicate {
		t.Errorf("expected StatusDuplicate, got %v", status)
	}

	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}

	if !strings.Contains(err.Error(), "duplicate transaction detected") {
		t.Errorf("expected error to contain 'duplicate transaction detected', got '%v'", err)
	}
}

// =========================================================================
// TEST CASE 7: repo.Create returns a generic DB error (not a sentinel)
//
//	-> falls into internal error branch -> StatusRejected + masked error
//
// =========================================================================
func TestProcessOne_RepoRandomError_ShouldReturnRejectedWithMaskedError(t *testing.T) {
	m := validMovement()

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			return nil
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			return errors.New("pq: deadlock detected")
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "cannot process movement, please try again"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected masked error '%s', got '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 8: MovementTypeIn -> quantityChange must be positive (+Quantity)
// =========================================================================
func TestProcessOne_MovementTypeIn_ShouldPassPositiveQuantityToAdjustStock(t *testing.T) {
	m := validMovement()
	m.Type = entity.MovementTypeIn
	m.Quantity = 25

	var capturedQuantity int32

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			capturedQuantity = quantityChange
			return nil
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != entity.StatusAccepted {
		t.Errorf("expected StatusAccepted, got %v", status)
	}

	if capturedQuantity != 25 {
		t.Errorf("expected quantityChange=+25 for MovementTypeIn, got %d", capturedQuantity)
	}
}

// =========================================================================
// TEST CASE 9: MovementTypeOut -> quantityChange must be negative (-Quantity)
// =========================================================================
func TestProcessOne_MovementTypeOut_ShouldPassNegativeQuantityToAdjustStock(t *testing.T) {
	m := validMovement()
	m.Type = entity.MovementTypeOut
	m.Quantity = 10

	var capturedQuantity int32

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			capturedQuantity = quantityChange
			return nil
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != entity.StatusAccepted {
		t.Errorf("expected StatusAccepted, got %v", status)
	}

	if capturedQuantity != -10 {
		t.Errorf("expected quantityChange=-10 for MovementTypeOut, got %d", capturedQuantity)
	}
}

// =========================================================================
// TEST CASE 10: MovementTypeAdjust -> quantityChange passed through as-is
// =========================================================================
func TestProcessOne_MovementTypeAdjust_ShouldPassQuantityDirectlyToAdjustStock(t *testing.T) {
	m := validMovement()
	m.Type = entity.MovementTypeAdjust
	m.Quantity = -5 // ADJUST allows negative values

	var capturedQuantity int32

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			capturedQuantity = quantityChange
			return nil
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != entity.StatusAccepted {
		t.Errorf("expected StatusAccepted, got %v", status)
	}

	if capturedQuantity != -5 {
		t.Errorf("expected quantityChange=-5 for MovementTypeAdjust, got %d", capturedQuantity)
	}
}

// =========================================================================
// TEST CASE 11: Happy path -> StatusAccepted, both AdjustStock and repo.Create called
// =========================================================================
func TestProcessOne_HappyPath_ShouldReturnAcceptedAndCallBothDownstreams(t *testing.T) {
	m := validMovement()

	adjustStockCalled := false
	repoCreateCalled := false

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			adjustStockCalled = true
			return nil
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			repoCreateCalled = true
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != entity.StatusAccepted {
		t.Errorf("expected StatusAccepted, got %v", status)
	}

	if !adjustStockCalled {
		t.Error("expected itemService.AdjustStock to be called")
	}

	if !repoCreateCalled {
		t.Error("expected movementRepo.Create to be called")
	}
}

// =========================================================================
// TEST CASE 12: AdjustStock fails -> transaction aborts early -> repo.Create must NOT be called
// =========================================================================
func TestProcessOne_AdjustStockFails_ShouldNotCallRepoCreate(t *testing.T) {
	m := validMovement()

	repoCreateCalled := false

	mockItemSvc := &mockItemService{
		adjustStockFn: func(ctx context.Context, itemID int32, quantityChange int32) error {
			return itemEntity.ErrInsufficientStock
		},
	}
	mockRepo := &mockMovementRepo{
		createFn: func(ctx context.Context, mv *entity.Movement) error {
			repoCreateCalled = true
			return nil
		},
	}

	svc := newMovementService(mockRepo, mockItemSvc, &mockTxManager{}, &mockWorkerPool{}, &mockLogger{})

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected StatusRejected, got %v", status)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repoCreateCalled {
		t.Error("expected repo.Create NOT to be called when AdjustStock fails")
	}
}
