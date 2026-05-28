package service

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	"strings"
	"testing"
)

// =========================================================================
// TEST CASE 1: GetItemForUpdate returns nil item (not found) -> Should return NotFound
// =========================================================================
func TestAdjustStock_ItemNotFound_ShouldReturnNotFound(t *testing.T) {
	var targetID int32 = 999
	var qtyChange int32 = -5

	mockRepo := &mockItemRepo{
		// repo returns nil, nil to signal "row not found" without a DB error
		// service then checks item == nil and returns ErrNotFound
		getItemForUpdateFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			return nil, nil
		},
	}

	svc := newService(mockRepo, &mockLogger{})

	err := svc.AdjustStock(context.Background(), targetID, qtyChange)

	if err == nil {
		t.Fatalf("expected not found error, got nil")
	}

	if !strings.Contains(err.Error(), "item not found") {
		t.Errorf("expected error message 'item not found', got: '%v'", err)
	}
}

// =========================================================================
// TEST CASE 2: Lock Item Fail (DB error) -> Should return Internal Error
// =========================================================================
func TestAdjustStock_LockItemFail_ShouldReturnInternalError(t *testing.T) {
	var targetID int32 = 101
	dbErr := errors.New("database connection lost during SELECT FOR UPDATE")

	mockRepo := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			return nil, dbErr
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	err := svc.AdjustStock(context.Background(), targetID, 10)

	if err == nil {
		t.Fatalf("expected internal error during lock, got nil")
	}

	expectedMsg := "failed to lock item for adjustment"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be masked with '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 3: Insufficient Stock (Negative result) -> Should return BadRequest
// =========================================================================
func TestAdjustStock_InsufficientStock_ShouldReturnBadRequest(t *testing.T) {
	var targetID int32 = 102
	var qtyChange int32 = -20

	mockRepo := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			return &entity.Item{
				CurrentStock: 15,
			}, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	err := svc.AdjustStock(context.Background(), targetID, qtyChange)

	if err == nil {
		t.Fatalf("expected validation error for negative stock, got nil")
	}

	if !strings.Contains(err.Error(), "insufficient stock") {
		t.Errorf("expected error to contain 'insufficient stock', got '%v'", err)
	}
}

// =========================================================================
// TEST CASE 4: Update Stock Fail -> Should return Internal Error
// =========================================================================
func TestAdjustStock_UpdateStockFail_ShouldReturnInternalError(t *testing.T) {
	var targetID int32 = 103
	var qtyChange int32 = -5

	mockRepo := &mockItemRepo{
		getItemForUpdateFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			return &entity.Item{
				CurrentStock: 50,
			}, nil
		},
		updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
			return errors.New("update command timeout")
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	err := svc.AdjustStock(context.Background(), targetID, qtyChange)

	if err == nil {
		t.Fatalf("expected internal error during update, got nil")
	}

	expectedMsg := "failed to update stock"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be masked with '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 5: Success Flow (Increase & Decrease) -> Should Return Nil Error
// =========================================================================
func TestAdjustStock_Success_ShouldUpdateStockCorrectly(t *testing.T) {
	tests := []struct {
		name          string
		initialStock  int32
		change        int32
		expectedStock int32
	}{
		{
			name:          "Increase stock successfully",
			initialStock:  10,
			change:        5,
			expectedStock: 15,
		},
		{
			name:          "Decrease stock successfully",
			initialStock:  10,
			change:        -4,
			expectedStock: 6,
		},
		{
			name:          "Decrease stock down to exactly zero",
			initialStock:  10,
			change:        -10,
			expectedStock: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetID int32 = 200
			updateStockCalled := false

			mockRepo := &mockItemRepo{
				getItemForUpdateFn: func(ctx context.Context, id int32) (*entity.Item, error) {
					return &entity.Item{
						CurrentStock: tt.initialStock,
					}, nil
				},
				updateStockFn: func(ctx context.Context, itemID int32, newStock int32) error {
					updateStockCalled = true
					if newStock != tt.expectedStock {
						t.Errorf("expected new stock to be %d, got %d", tt.expectedStock, newStock)
					}
					return nil
				},
			}
			mockLog := &mockLogger{}

			svc := newService(mockRepo, mockLog)

			err := svc.AdjustStock(context.Background(), targetID, tt.change)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !updateStockCalled {
				t.Errorf("expected repo.UpdateStock to be called")
			}
		})
	}
}
