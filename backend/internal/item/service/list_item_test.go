package service

import (
	"context"
	"errors"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"strings"
	"testing"
)

// =========================================================================
// TEST CASE 1: Repository Error -> Should return Internal Error
// =========================================================================
func TestListItem_RepositoryError_ShouldReturnInternalError(t *testing.T) {
	filter := &entity.ItemFilter{}
	paging := &core.Pagination{Page: 1, Limit: 10}
	dbErr := errors.New("sql: syntax error or table not found")

	mockRepo := &mockItemRepo{
		listItemFn: func(ctx context.Context, f *entity.ItemFilter, p *core.Pagination) ([]entity.Item, error) {
			return nil, dbErr
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.ListItem(context.Background(), filter, paging)

	if result != nil {
		t.Errorf("expected result to be nil upon repository failure, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected internal error, got nil")
	}

	expectedMsg := "failed to fetch items"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be safely masked, expected: '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 2: Success Flow with Empty Result -> Should return Empty Slice and Nil Error
// =========================================================================
func TestListItem_Success_EmptyResult(t *testing.T) {
	filter := &entity.ItemFilter{}
	paging := &core.Pagination{Page: 1, Limit: 10}

	mockRepo := &mockItemRepo{
		listItemFn: func(ctx context.Context, f *entity.ItemFilter, p *core.Pagination) ([]entity.Item, error) {
			return []entity.Item{}, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.ListItem(context.Background(), filter, paging)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected an empty slice, got nil")
	}

	if len(result) != 0 {
		t.Errorf("expected slice length to be 0, got %d", len(result))
	}
}

// =========================================================================
// TEST CASE 3: Success Flow with Data -> Should return Items list correctly
// =========================================================================
func TestListItem_Success_ShouldReturnItemsList(t *testing.T) {
	searchName := "iPhone"
	filter := &entity.ItemFilter{Name: &searchName}
	paging := &core.Pagination{Page: 1, Limit: 2}

	expectedItems := []entity.Item{
		{Name: "iPhone 15 Pro", SKU: "SKU-IPHONE-15"},
		{Name: "iPhone 14 Plus", SKU: "SKU-IPHONE-14"},
	}

	mockRepo := &mockItemRepo{
		listItemFn: func(ctx context.Context, f *entity.ItemFilter, p *core.Pagination) ([]entity.Item, error) {
			if f.Name != filter.Name || p.Page != paging.Page || p.Limit != paging.Limit {
				t.Errorf("repo called with unexpected arguments")
			}
			return expectedItems, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newService(mockRepo, mockLog)

	result, err := svc.ListItem(context.Background(), filter, paging)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	if result[0].SKU != expectedItems[0].SKU || result[1].Name != expectedItems[1].Name {
		t.Errorf("returned items data does not match repository mock data")
	}
}