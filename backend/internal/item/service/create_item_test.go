package service

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	"strings"
	"testing"
)

// =========================================================================
// TEST CASE 1: Duplicate SKU -> Should return Conflict Error
// =========================================================================
func TestCreateItem_DuplicateSKU_ShouldReturnConflict(t *testing.T) {
	input := entity.CreateItemRequest{
		Name: "Laptop Dell XPS 15",
		SKU:  "SKU-DUP-001",
	}

	mockRepo := &mockItemRepo{
		createItemFn: func(ctx context.Context, item entity.Item) (*entity.Item, error) {
			return nil, entity.ErrItemDuplicated
		},
	}

	svc := newService(mockRepo, &mockLogger{})

	result, err := svc.CreateItem(context.Background(), input)

	if result != nil {
		t.Errorf("expected result to be nil upon duplicate SKU, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected conflict error, got nil")
	}

	if !strings.Contains(err.Error(), "duplicate SKU detected") {
		t.Errorf("expected error to contain 'duplicate SKU detected', got '%v'", err)
	}
}

// =========================================================================
// TEST CASE 2: Repository random error -> Should return masked Internal Error
// =========================================================================
func TestCreateItem_RepositoryError_ShouldReturnInternalError(t *testing.T) {
	input := entity.CreateItemRequest{
		Name: "Mechanical Keyboard",
		SKU:  "SKU-KEY-999",
	}

	dbErr := errors.New("connection timeout or database down")

	mockRepo := &mockItemRepo{
		createItemFn: func(ctx context.Context, item entity.Item) (*entity.Item, error) {
			return nil, dbErr
		},
	}

	svc := newService(mockRepo, &mockLogger{})

	result, err := svc.CreateItem(context.Background(), input)

	if result != nil {
		t.Errorf("expected result to be nil upon repository failure, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected internal error, got nil")
	}

	expectedMsg := "failed to create item, please try again"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be safely masked with '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 3: Success -> Should return created Item with repo-assigned ID
// =========================================================================
func TestCreateItem_Success_ShouldReturnCreatedItem(t *testing.T) {
	input := entity.CreateItemRequest{
		Name:              "Sony WH-1000XM4",
		SKU:               "SKU-SONY-04",
		LowStockThreshold: 15,
	}

	mockRepo := &mockItemRepo{
		createItemFn: func(ctx context.Context, item entity.Item) (*entity.Item, error) {
			item.ID = 101
			return &item, nil
		},
	}

	svc := newService(mockRepo, &mockLogger{})

	result, err := svc.CreateItem(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected created item, got nil")
	}

	if result.ID != 101 {
		t.Errorf("expected ID=101, got %d", result.ID)
	}

	if result.Name != input.Name {
		t.Errorf("expected Name='%s', got '%s'", input.Name, result.Name)
	}

	if result.SKU != input.SKU {
		t.Errorf("expected SKU='%s', got '%s'", input.SKU, result.SKU)
	}

	if result.LowStockThreshold != input.LowStockThreshold {
		t.Errorf("expected LowStockThreshold=%d, got %d", input.LowStockThreshold, result.LowStockThreshold)
	}
}

// =========================================================================
// TEST CASE 4: Success -> Should map CreateItemRequest fields to Item correctly
// =========================================================================
func TestCreateItem_Success_ShouldMapRequestFieldsCorrectly(t *testing.T) {
	input := entity.CreateItemRequest{
		Name:              "Standing Desk",
		SKU:               "SKU-DESK-01",
		LowStockThreshold: 5,
	}

	var capturedItem entity.Item

	mockRepo := &mockItemRepo{
		createItemFn: func(ctx context.Context, item entity.Item) (*entity.Item, error) {
			capturedItem = item
			item.ID = 55
			return &item, nil
		},
	}

	svc := newService(mockRepo, &mockLogger{})

	_, err := svc.CreateItem(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify service maps request -> entity.Item correctly before passing to repo
	if capturedItem.Name != input.Name {
		t.Errorf("expected captured Name='%s', got '%s'", input.Name, capturedItem.Name)
	}

	if capturedItem.SKU != input.SKU {
		t.Errorf("expected captured SKU='%s', got '%s'", input.SKU, capturedItem.SKU)
	}

	if capturedItem.LowStockThreshold != input.LowStockThreshold {
		t.Errorf("expected captured LowStockThreshold=%d, got %d", input.LowStockThreshold, capturedItem.LowStockThreshold)
	}
}
