package service

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	"strings"
	"testing"
)

// =========================================================================
// TEST CASE 1: Repository Error -> Should return Internal Error
// =========================================================================
func TestListLowStockItems_RepositoryError_ShouldReturnInternalError(t *testing.T) {
	dbErr := errors.New("database connection timeout")

	mockRepo := &mockItemRepo{
		listLowStockFn: func(ctx context.Context) ([]*entity.Item, error) {
			return nil, dbErr
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.ListLowStockItems(context.Background())

	if result != nil {
		t.Errorf("expected result to be nil upon repository failure, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected internal error, got nil")
	}

	// Đảm bảo thông tin lỗi thô từ dbErr được bọc qua common.ErrInternal (theo logic err.Error())
	if !strings.Contains(err.Error(), dbErr.Error()) {
		t.Errorf("expected error message to contain '%s', got '%v'", dbErr.Error(), err)
	}
}

// =========================================================================
// TEST CASE 2: Success Flow with Empty Result -> Should return Empty Slice
// =========================================================================
func TestListLowStockItems_Success_EmptyResult(t *testing.T) {
	mockRepo := &mockItemRepo{
		listLowStockFn: func(ctx context.Context) ([]*entity.Item, error) {
			return []*entity.Item{}, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.ListLowStockItems(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected empty slice, got nil")
	}

	if len(result) != 0 {
		t.Errorf("expected slice length to be 0, got %d", len(result))
	}
}

// =========================================================================
// TEST CASE 3: Success Flow with Data -> Should return Low Stock Items List
// =========================================================================
func TestListLowStockItems_Success_ShouldReturnItemsList(t *testing.T) {
	expectedItems := []*entity.Item{
		{Name: "Mì ăn liền", SKU: "SKU-NOODLE", CurrentStock: 2, LowStockThreshold: 10},
		{Name: "Nước khoáng", SKU: "SKU-WATER", CurrentStock: 5, LowStockThreshold: 20},
	}

	mockRepo := &mockItemRepo{
		listLowStockFn: func(ctx context.Context) ([]*entity.Item, error) {
			return expectedItems, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.ListLowStockItems(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 low stock items, got %d", len(result))
	}

	if result[0].SKU != expectedItems[0].SKU || result[1].CurrentStock != expectedItems[1].CurrentStock {
		t.Errorf("returned low stock items do not match repository mock data")
	}
}