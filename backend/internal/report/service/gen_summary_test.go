package service

import (
	"context"
	"errors"
	"testing"
	"time"

	reportEntity "inventory-movement-processing/internal/report/entity"
)

func TestReportService_GenerateDailySummary(t *testing.T) {
	ctx := context.Background()

	testDate := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)

	mockErr := errors.New("mock unexpected error")

	mockSummaries := []*reportEntity.DailyItemSummary{
		{ItemID: 1, TotalIn: 10, TotalOut: 5},
	}

	tests := []struct {
		name        string
		mockMove    *mockMovementUseCase
		mockRepo    *mockReportRepo
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Case 1: Success with aggregated data",
			mockMove: &mockMovementUseCase{
				aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
					return mockSummaries, nil
				},
			},
			mockRepo: &mockReportRepo{
				upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
					if len(data) != 1 {
						t.Errorf("expected 1 item, got %d", len(data))
					}

					return nil
				},
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "Case 2: Success with empty aggregated data",
			mockMove: &mockMovementUseCase{
				aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
					return []*reportEntity.DailyItemSummary{}, nil
				},
			},
			mockRepo: &mockReportRepo{
				upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
					t.Errorf("upsert should not be called when there is no data")
					return nil
				},
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "Case 3: Failed to save aggregated data",
			mockMove: &mockMovementUseCase{
				aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
					return mockSummaries, nil
				},
			},
			mockRepo: &mockReportRepo{
				upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
					return mockErr
				},
			},
			wantErr:     true,
			expectedErr: mockErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockItemRepo := &mockItemRepo{}
			mockCache := &mockCache{}
			mockConfig := &mockConfig{}
			mockLogger := &mockLogger{}

			svc := newService(
				tt.mockRepo,
				tt.mockMove,
				mockItemRepo,
				mockCache,
				mockConfig,
				mockLogger,
			)

			err := svc.GenerateDailySummary(ctx, testDate)

			// VERIFY
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDailySummary() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("GenerateDailySummary() expected error %v, got %v",
					tt.expectedErr,
					err,
				)
			}
		})
	}
}
