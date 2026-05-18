package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/item/service/mocks"
	"inventory-movement-processing/pkg/core"
	"testing"
)

func TestListItem(t *testing.T) {
	ctx := context.Background()

	fakeItems := []entity.Item{
		{Name: "Item 1", SKU: "SKU-001", CurrentStock: 10, LowStockThreshold: 5},
		{Name: "Item 2", SKU: "SKU-002", CurrentStock: 20, LowStockThreshold: 3},
	}

	tests := []struct {
		name        string
		filter      *entity.ItemFilter
		repoReturn  []entity.Item
		repoErr     error
		expectedErr error
	}{
		{
			name:        "Success: repo returns list of items",
			filter:      &entity.ItemFilter{},
			repoReturn:  fakeItems,
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Internal Error: repo fails, should wrap into ErrInternal",
			filter:      &entity.ItemFilter{},
			repoReturn:  nil,
			repoErr:     errors.New("db connection lost"),
			expectedErr: common.ErrInternal("cannot list items"),
		},
		{
			name:       "Success: repo returns empty list",
			filter:     &entity.ItemFilter{},
			repoReturn: []entity.Item{},
			repoErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewItemService(&mocks.ItemRepository{
				ListItemFn: func(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
					if paging != nil && tt.repoErr == nil {
						paging.Total = len(tt.repoReturn)
					}
					return tt.repoReturn, tt.repoErr
				},
			})

			paging := &core.Pagination{Page: 1, Limit: 10}

			result, err := svc.ListItem(ctx, tt.filter, paging)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if len(result) != len(tt.repoReturn) {
					t.Errorf("Expected %d items, got %d", len(tt.repoReturn), len(result))
				}
				if paging.Total != len(tt.repoReturn) {
					t.Errorf("Expected paging.Total = %d, got %d", len(tt.repoReturn), paging.Total)
				}
				for i, item := range result {
					if item.Name != tt.repoReturn[i].Name {
						t.Errorf("Expected item[%d].Name = %q, got %q", i, tt.repoReturn[i].Name, item.Name)
					}
					if item.SKU != tt.repoReturn[i].SKU {
						t.Errorf("Expected item[%d].SKU = %q, got %q", i, tt.repoReturn[i].SKU, item.SKU)
					}
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedErr)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got: %+v", result)
				}

				var appErr *common.AppError
				if !errors.As(err, &appErr) {
					t.Fatalf("Expected *common.AppError, got %T: %v", err, err)
				}

				var expectedAppErr *common.AppError
				errors.As(tt.expectedErr, &expectedAppErr)

				if appErr.StatusCode != expectedAppErr.StatusCode {
					t.Errorf("Expected StatusCode = %d, got %d", expectedAppErr.StatusCode, appErr.StatusCode)
				}
				if appErr.Message != expectedAppErr.Message {
					t.Errorf("Expected Message = %q, got %q", expectedAppErr.Message, appErr.Message)
				}
			}
		})
	}
}
