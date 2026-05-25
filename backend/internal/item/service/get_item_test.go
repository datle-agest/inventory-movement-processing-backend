package service

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	"strings"
	"testing"
)

// =========================================================================
// TEST CASE 1: GetItem returns nil item (not found) -> Should return NotFound
// =========================================================================
func TestGetItem_ItemNotFound_ShouldReturnNotFound(t *testing.T) {
	var targetID int32 = 404

	mockRepo := &mockItemRepo{
		// repo returns nil, nil to signal "row not found" without a DB error
		// service then checks item == nil and returns ErrNotFound
		getItemFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			return nil, nil
		},
	}

	svc := newService(mockRepo, &mockLogger{})

	result, err := svc.GetItem(context.Background(), targetID)

	if result != nil {
		t.Errorf("expected result to be nil, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected not found error, got nil")
	}

	if !strings.Contains(err.Error(), "item not found") {
		t.Errorf("expected error message 'item not found', got: '%v'", err)
	}
}

// =========================================================================
// TEST CASE 2: Repository Error -> Should return Internal Error
// =========================================================================
func TestGetItem_RepositoryError_ShouldReturnInternalError(t *testing.T) {
	var targetID int32 = 500
	dbErr := errors.New("sql: database connection closed")

	mockRepo := &mockItemRepo{
		getItemFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			return nil, dbErr
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.GetItem(context.Background(), targetID)

	if result != nil {
		t.Errorf("expected result to be nil upon repository failure, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected internal server error, got nil")
	}

	expectedMsg := "internal server error"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be safely masked, expected: '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 3: Success Flow -> Should return Item data
// =========================================================================
func TestGetItem_Success_ShouldReturnItem(t *testing.T) {
	var targetID int32 = 123
	expectedItem := &entity.Item{
		Name:         "MacBook Pro M3",
		SKU:          "SKU-MAC-M3",
		CurrentStock: 25,
	}
	expectedItem.ID = targetID

	mockRepo := &mockItemRepo{
		getItemFn: func(ctx context.Context, id int32) (*entity.Item, error) {
			if id != targetID {
				t.Errorf("expected repo to be called with id %d, got %d", targetID, id)
			}
			return expectedItem, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.GetItem(context.Background(), targetID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected to get item data, got nil")
	}

	if result.ID != targetID {
		t.Errorf("expected item ID to be %d, got %d", targetID, result.ID)
	}

	if result.SKU != expectedItem.SKU || result.Name != expectedItem.Name {
		t.Errorf("returned item properties do not match expected data")
	}
}