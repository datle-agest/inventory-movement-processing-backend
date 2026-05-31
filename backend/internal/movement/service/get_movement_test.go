package service

import (
	"context"
	"errors"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/core"
	"strings"
	"testing"
)

// =========================================================================
// TEST CASE 1: Invalid Item ID (<= 0) -> Should return BadRequest
// =========================================================================
func TestGetMovementsByItemID_InvalidItemID_ShouldReturnBadRequest(t *testing.T) {
	invalidIDs := []int{0, -1, -99}

	for _, id := range invalidIDs {
		mockRepo := &mockMovementRepo{}
		mockLog := &mockLogger{}

		svc := newMovementService(mockRepo, nil, nil, nil, mockLog)
		paging := &core.Pagination{Page: 1, Limit: 10}

		result, err := svc.GetMovementsByItemID(context.Background(), id, paging)

		if result != nil {
			t.Errorf("expected result to be nil for invalid id %d, got %v", id, result)
		}

		if err == nil {
			t.Fatalf("expected bad request error for id %d, got nil", id)
		}

		expectedMsg := "invalid item id for history lookup"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error message to be '%s', got: '%v'", expectedMsg, err)
		}
	}
}

// =========================================================================
// TEST CASE 2: Repository Error -> Should return Internal Error
// =========================================================================
func TestGetMovementsByItemID_RepositoryError_ShouldReturnInternalError(t *testing.T) {
	var targetItemID int = 42
	paging := &core.Pagination{Page: 1, Limit: 10}
	dbErr := errors.New("sql: database connection closed unexpectedly")

	mockRepo := &mockMovementRepo{
		getMovementsByItemIDFn: func(ctx context.Context, itemId int, p *core.Pagination) ([]*movementEntity.Movement, error) {
			return nil, dbErr
		},
	}

	mockItem := &mockItemService{
		getItemFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{}, nil
		},
	}

	mockLog := &mockLogger{}

	svc := newMovementService(mockRepo, mockItem, nil, nil, mockLog)

	result, err := svc.GetMovementsByItemID(context.Background(), targetItemID, paging)

	if result != nil {
		t.Errorf("expected result to be nil upon repository failure, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected internal error, got nil")
	}

	expectedMsg := "failed to fetch movement history"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be safely masked with '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 3: Success Flow -> Should return Movements List Correctly
// =========================================================================
func TestGetMovementsByItemID_Success_ShouldReturnMovementsList(t *testing.T) {
	var targetItemID int = 100
	paging := &core.Pagination{Page: 1, Limit: 2}

	expectedMovements := []*movementEntity.Movement{
		{ItemID: 100, Quantity: 50, Type: "IN"},
		{ItemID: 100, Quantity: 20, Type: "OUT"},
	}

	repoCalled := false

	mockRepo := &mockMovementRepo{
		getMovementsByItemIDFn: func(ctx context.Context, itemId int, p *core.Pagination) ([]*movementEntity.Movement, error) {
			repoCalled = true

			if itemId != targetItemID {
				t.Errorf("expected repository to be called with item id %d, got %d", targetItemID, itemId)
			}
			if p.Page != paging.Page || p.Limit != paging.Limit {
				t.Errorf("expected repository to be called with correct pagination config")
			}

			return expectedMovements, nil
		},
	}

	mockItem := &mockItemService{
		getItemFn: func(ctx context.Context, id int32) (*itemEntity.Item, error) {
			return &itemEntity.Item{}, nil
		},
	}

	mockLog := &mockLogger{}

	svc := newMovementService(mockRepo, mockItem, nil, nil, mockLog)

	result, err := svc.GetMovementsByItemID(context.Background(), targetItemID, paging)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repoCalled {
		t.Fatalf("expected repository's GetMovementsByItemID to be called")
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 movement records, got %d", len(result))
	}

	if result[0].Type != "IN" || result[1].Quantity != 20 {
		t.Errorf("returned movement history records do not match mock repository response")
	}
}
