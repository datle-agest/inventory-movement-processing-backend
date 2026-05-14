package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/item/service/mocks"
	"testing"
)

func TestCreateItem(t *testing.T) {
	ctx := context.Background()

	validItem := entity.Item{
		Name:              "Test Item",
		SKU:               "SKU-001",
		CurrentStock:      10,
		LowStockThreshold: 5,
	}

	createdItem := &entity.Item{
		Name:              "Test Item",
		SKU:               "SKU-001",
		CurrentStock:      10,
		LowStockThreshold: 5,
	}

	tests := []struct {
		name        string
		input       entity.Item
		repoReturn  *entity.Item
		repoErr     error
		expectedErr error
	}{
		{
			name:        "Success: valid item created successfully",
			input:       validItem,
			repoReturn:  createdItem,
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Bad Request: empty item name",
			input:       entity.Item{SKU: "SKU-001"},
			repoReturn:  nil,
			repoErr:     nil,
			expectedErr: common.ErrBadRequest("item name is required"),
		},
		{
			name:        "Internal Error: repo fails on create",
			input:       validItem,
			repoReturn:  nil,
			repoErr:     errors.New("db connection lost"),
			expectedErr: common.ErrInternal("cannot create item"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewItemService(&mocks.ItemRepository{
				CreateItemFn: func(ctx context.Context, item entity.Item) (*entity.Item, error) {
					return tt.repoReturn, tt.repoErr
				},
			})

			result, err := svc.CreateItem(ctx, tt.input)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected created item, got nil")
				}
				if result.Name != tt.repoReturn.Name {
					t.Errorf("Expected Name = %q, got %q", tt.repoReturn.Name, result.Name)
				}
				if result.SKU != tt.repoReturn.SKU {
					t.Errorf("Expected SKU = %q, got %q", tt.repoReturn.SKU, result.SKU)
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
