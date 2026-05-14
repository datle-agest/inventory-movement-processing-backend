package service

import (
	"context"
	"errors"
	"inventory-movement-processing/common"
	"inventory-movement-processing/internal/item/service/mocks"
	"testing"

	"gorm.io/gorm"
)

func TestDeleteItem(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int
		repoErr     error
		expectedErr error
	}{
		{
			name:        "Success: item deleted successfully",
			id:          1,
			repoErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "Not Found: repo returns gorm.ErrRecordNotFound, should wrap into ErrNotFound",
			id:          999,
			repoErr:     gorm.ErrRecordNotFound,
			expectedErr: common.ErrNotFound("item not found"),
		},
		{
			name:        "Internal Error: repo returns unexpected error, should wrap into ErrInternal",
			id:          1,
			repoErr:     errors.New("db connection lost"),
			expectedErr: common.ErrInternal("cannot delete item"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewItemService(&mocks.ItemRepository{
				DeleteItemFn: func(ctx context.Context, id int) error {
					return tt.repoErr
				},
			})

			err := svc.DeleteItem(ctx, tt.id)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectedErr)
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
