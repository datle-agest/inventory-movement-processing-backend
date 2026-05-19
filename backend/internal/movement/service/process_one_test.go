package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
)

func TestProcessOne_InvalidInput(t *testing.T) {
	// ItemID = 0 is invalid
	m := &movementEntity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       0,
		Type:         movementEntity.MovementTypeIn,
		Quantity:     10,
		MovementTime: time.Now(),
	}

	svc := NewMovementService(nil, &mockItemRepo{}, &mockTxManager{}, nil)

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected status %s, got %s", entity.StatusRejected, status)
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected common.AppError, got %T", err)
	}

	if appErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected HTTP 400, got %d", appErr.StatusCode)
	}
}

func TestProcessOne_Success(t *testing.T) {
	m := &movementEntity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       1,
		Type:         movementEntity.MovementTypeIn,
		Quantity:     10,
		MovementTime: time.Now(),
	}

	var processCalled bool

	mr := &mockMovementRepo{
		processMovementFn: func(ctx context.Context, gotM *movementEntity.Movement) error {
			processCalled = true

			if gotM.ExternalID != m.ExternalID {
				t.Errorf("expected external ID %s, got %s", m.ExternalID, gotM.ExternalID)
			}

			return nil
		},
	}

	// Mock item exists
	ir := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{
				CurrentStock: 50,
			}, nil
		},
	}

	svc := NewMovementService(mr, ir, &mockTxManager{}, nil)

	status, err := svc.ProcessOne(context.Background(), m)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != entity.StatusAccepted {
		t.Errorf("expected status %s, got %s", entity.StatusAccepted, status)
	}

	if !processCalled {
		t.Error("expected ProcessMovement to be called")
	}
}

func TestProcessOne_ItemNotFound(t *testing.T) {
	m := &movementEntity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       1,
		Type:         movementEntity.MovementTypeIn,
		Quantity:     10,
		MovementTime: time.Now(),
	}

	ir := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return nil, itemEntity.ErrItemNotFound
		},
	}

	svc := NewMovementService(nil, ir, &mockTxManager{}, nil)

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected status %s, got %s", entity.StatusRejected, status)
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected common.AppError, got %T", err)
	}

	if appErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected HTTP 404, got %d", appErr.StatusCode)
	}

	if appErr.Message != "inventory item not found" {
		t.Errorf("expected error message 'inventory item not found', got '%s'", appErr.Message)
	}
}

func TestProcessOne_InsufficientStock(t *testing.T) {
	m := &movementEntity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       1,
		Type:         movementEntity.MovementTypeOut,
		Quantity:     10,
		MovementTime: time.Now(),
	}

	ir := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			// current stock = 5, but request out = 10
			return &itemEntity.Item{
				CurrentStock: 5,
			}, nil
		},
	}

	svc := NewMovementService(nil, ir, &mockTxManager{}, nil)

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected status %s, got %s", entity.StatusRejected, status)
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected common.AppError, got %T", err)
	}

	if appErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected HTTP 400, got %d", appErr.StatusCode)
	}

	if appErr.Message != "insufficient stock" {
		t.Errorf("expected error message 'insufficient stock', got '%s'", appErr.Message)
	}
}

func TestProcessOne_DuplicateMovement(t *testing.T) {
	m := &movementEntity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       1,
		Type:         movementEntity.MovementTypeIn,
		Quantity:     10,
		MovementTime: time.Now(),
	}

	mr := &mockMovementRepo{
		createFn: func(ctx context.Context, gotM *movementEntity.Movement) error {
			return itemEntity.ErrDuplicateMovement
		},
	}

	ir := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{
				CurrentStock: 50,
			}, nil
		},
	}

	svc := NewMovementService(mr, ir, &mockTxManager{}, nil)

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusDuplicate {
		t.Errorf("expected status %s, got %s", entity.StatusDuplicate, status)
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected common.AppError, got %T", err)
	}

	if appErr.StatusCode != http.StatusConflict {
		t.Errorf("expected HTTP 409, got %d", appErr.StatusCode)
	}

	if appErr.Message != "duplicate external_id" {
		t.Errorf("expected error message 'duplicate external_id', got '%s'", appErr.Message)
	}
}

func TestProcessOne_InternalError(t *testing.T) {
	m := &movementEntity.Movement{
		ExternalID:   "EXT-001",
		ItemID:       1,
		Type:         movementEntity.MovementTypeIn,
		Quantity:     10,
		MovementTime: time.Now(),
	}

	mr := &mockMovementRepo{
		processMovementFn: func(ctx context.Context, gotM *movementEntity.Movement) error {
			return errors.New("db connection failure")
		},
	}

	ir := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{
				CurrentStock: 50,
			}, nil
		},
	}

	svc := NewMovementService(mr, ir, &mockTxManager{}, nil)

	status, err := svc.ProcessOne(context.Background(), m)

	if status != entity.StatusRejected {
		t.Errorf("expected status %s, got %s", entity.StatusRejected, status)
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected common.AppError, got %T", err)
	}

	if appErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected HTTP 500, got %d", appErr.StatusCode)
	}

	if appErr.Message != "cannot process movement" {
		t.Errorf("expected error message 'cannot process movement', got '%s'", appErr.Message)
	}
}
