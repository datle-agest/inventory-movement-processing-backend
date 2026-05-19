package service

import (
	"context"
	"errors"
	"testing"

	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/entity"

	"inventory-movement-processing/internal/item/service/mocks"

	"gorm.io/gorm"
)

func TestGetItem(t *testing.T) {
	ctx := context.Background()

	fakeItem := &entity.Item{
		Name:              "Test Item",
		SKU:               "SKU-001",
		CurrentStock:      10,
		LowStockThreshold: 5,
	}

	tests := []struct {
		name        string
		id          int32
		repoReturn  *entity.Item
		repoErr     error
		expectedErr error
	}{
		{
			name:        "Success: repo returns a valid item",
			id:          1,
			repoReturn:  fakeItem,
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Not Found: repo returns gorm.ErrRecordNotFound, should wrap into ErrNotFound",
			id:          999,
			repoReturn:  nil,
			repoErr:     gorm.ErrRecordNotFound,
			expectedErr: common.ErrNotFound("item not found"),
		},
		{
			name:        "Internal Error: repo returns unexpected error, should wrap into ErrInternal",
			id:          1,
			repoReturn:  nil,
			repoErr:     errors.New("db connection lost"),
			expectedErr: common.ErrInternal("cannot get item"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := &mocks.ItemRepository{
				GetItemFn: func(ctx context.Context, id int32) (*entity.Item, error) {
					return tt.repoReturn, tt.repoErr
				},
			}

			svc := NewItemService(mockRepo, &mocks.Logger{})

			item, err := svc.GetItem(ctx, tt.id)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if item == nil {
					t.Fatal("Expected item, got nil")
				}
				if item.Name != tt.repoReturn.Name {
					t.Errorf("Expected Name = %q, got %q", tt.repoReturn.Name, item.Name)
				}
				if item.SKU != tt.repoReturn.SKU {
					t.Errorf("Expected SKU = %q, got %q", tt.repoReturn.SKU, item.SKU)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedErr)
				}
				if item != nil {
					t.Errorf("Expected nil item on error, got: %+v", item)
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
