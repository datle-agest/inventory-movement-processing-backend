package service

import (
	"context"
	"errors"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"strings"
	"testing"
	"time"
)

// =========================================================================
// TEST CASE 1: Repository Error -> Should return Internal Error
// =========================================================================
func TestAggregateDailyItemSummaryFromMovement_RepositoryError_ShouldReturnInternalError(t *testing.T) {
	requestedDate := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	dbErr := errors.New("sql: database connection timed out during aggregation")

	mockRepo := &mockMovementRepo{
		aggregateDailyItemSummaryFromMovementFn: func(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error) {
			return nil, dbErr
		},
	}
	mockLog := &mockLogger{}

	svc := newMovementService(mockRepo, nil, nil, nil, mockLog)

	result, err := svc.AggregateDailyItemSummaryFromMovement(context.Background(), requestedDate)

	if result != nil {
		t.Errorf("expected result to be nil upon repository failure, got %v", result)
	}

	if err == nil {
		t.Fatalf("expected internal error, got nil")
	}

	expectedMsg := "failed to aggregate daily summary"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error message to be safely masked with '%s', got: '%v'", expectedMsg, err)
	}
}

// =========================================================================
// TEST CASE 2: Success Flow with Data -> Should Parse Time Range and Return Data Correctly
// =========================================================================
func TestAggregateDailyItemSummaryFromMovement_Success_ShouldReturnSummaries(t *testing.T) {
	requestedDate := time.Date(2026, time.May, 25, 14, 35, 18, 999, time.UTC)

	expectedStart := time.Date(2026, time.May, 25, 0, 0, 0, 0, time.UTC)
	expectedEnd := expectedStart.Add(24 * time.Hour)

	expectedSummaries := []*reportEntity.DailyItemSummary{
		{ItemID: 10, TotalIn: 100, TotalOut: 30, SummaryDate: expectedStart},
		{ItemID: 20, TotalIn: 50, TotalOut: 50, SummaryDate: expectedStart},
	}

	repoCalled := false

	mockRepo := &mockMovementRepo{
		aggregateDailyItemSummaryFromMovementFn: func(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error) {
			repoCalled = true

			if !start.Equal(expectedStart) {
				t.Errorf("expected start time to be %v, got %v", expectedStart, start)
			}

			if !end.Equal(expectedEnd) {
				t.Errorf("expected end time to be %v, got %v", expectedEnd, end)
			}

			return expectedSummaries, nil
		},
	}
	mockLog := &mockLogger{}

	svc := newMovementService(mockRepo, nil, nil, nil, mockLog)

	result, err := svc.AggregateDailyItemSummaryFromMovement(context.Background(), requestedDate)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repoCalled {
		t.Fatalf("expected repository's AggregateDailyItemSummaryFromMovement to be called")
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 summary items returned, got %d", len(result))
	}

	if result[0].ItemID != 10 || result[1].TotalIn != 50 {
		t.Errorf("returned aggregation data does not match mock repo response")
	}
}