package service

import (
	"context"
	"errors"
	"testing"
	"time"

	reportEntity "inventory-movement-processing/internal/report/entity"
)

func TestAggregateDailyItemSummaryFromMovement_Success(t *testing.T) {
	testDate := time.Date(2026, 5, 19, 15, 30, 0, 0, time.UTC)
	expectedStart := time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)
	expectedEnd := expectedStart.Add(24 * time.Hour)

	mockResult := []*reportEntity.DailyItemSummary{
		{ItemID: 1, TotalIn: 10, TotalOut: 5, TotalAdjust: 2},
	}

	var capturedStart, capturedEnd time.Time

	mr := &mockMovementRepo{
		aggregateDailyItemSummaryFromMovementFn: func(ctx context.Context, start, end time.Time) ([]*reportEntity.DailyItemSummary, error) {
			capturedStart = start
			capturedEnd = end
			return mockResult, nil
		},
	}

	svc := newMovementService(mr, nil)
	result, err := svc.AggregateDailyItemSummaryFromMovement(context.Background(), testDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !capturedStart.Equal(expectedStart) {
		t.Errorf("expected start time %v, got %v", expectedStart, capturedStart)
	}
	if !capturedEnd.Equal(expectedEnd) {
		t.Errorf("expected end time %v, got %v", expectedEnd, capturedEnd)
	}
	if len(result) != len(mockResult) {
		t.Fatalf("expected result size %d, got %d", len(mockResult), len(result))
	}
	if result[0].ItemID != 1 {
		t.Errorf("expected ItemID 1, got %d", result[0].ItemID)
	}
}

func TestAggregateDailyItemSummaryFromMovement_RepoError(t *testing.T) {
	wantErr := errors.New("repository aggregation error")

	mr := &mockMovementRepo{
		aggregateDailyItemSummaryFromMovementFn: func(ctx context.Context, start, end time.Time) ([]*reportEntity.DailyItemSummary, error) {
			return nil, wantErr
		},
	}

	svc := newMovementService(mr, nil)
	_, err := svc.AggregateDailyItemSummaryFromMovement(context.Background(), time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected error %v, got %v", wantErr, err)
	}
}
