package service

import (
	"context"
	"errors"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
)

func TestGetDailyReport_DefaultLimit(t *testing.T) {
	summaries := makeSummaries(10)

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error {
				return nil
			},
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		&mockMovementRepo{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		&mockItemRepo{
			listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) {
				return nil, nil
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	// limit <= 0 → default 5
	result, err := svc.GetDailyReport(context.Background(), today(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.TopItems) != 5 {
		t.Errorf("expected 5 items (default limit), got %d", len(result.TopItems))
	}
}

func TestGetDailyReport_CacheHit_Past(t *testing.T) {
	summaries := makeSummaries(10)
	cachedJSON := marshalSummaries(t, summaries)

	listCalled := false

	svc := newService(
		&mockReportRepo{
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				listCalled = true
				return nil, nil
			},
		},
		&mockMovementRepo{},
		nil,
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return cachedJSON, true, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	result, err := svc.GetDailyReport(context.Background(), yesterday(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listCalled {
		t.Error("should not hit DB on cache hit")
	}
	if len(result.TopItems) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.TopItems))
	}
	if result.LowStockItems != nil {
		t.Error("past date should not include low stock items")
	}
}

func TestGetDailyReport_CacheHit_Today_FetchesLowStock(t *testing.T) {
	summaries := makeSummaries(5)
	cachedJSON := marshalSummaries(t, summaries)

	lowStockItems := []*itemEntity.Item{{}, {}}

	svc := newService(
		&mockReportRepo{},
		&mockMovementRepo{},
		&mockItemRepo{
			listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) {
				return lowStockItems, nil
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return cachedJSON, true, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	result, err := svc.GetDailyReport(context.Background(), today(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LowStockItems) != 2 {
		t.Errorf("expected 2 low stock items, got %d", len(result.LowStockItems))
	}
}

func TestGetDailyReport_CacheHit_Today_LowStockError(t *testing.T) {
	summaries := makeSummaries(5)
	cachedJSON := marshalSummaries(t, summaries)
	wantErr := errors.New("low stock fetch failed")

	svc := newService(
		&mockReportRepo{},
		&mockMovementRepo{},
		&mockItemRepo{
			listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) {
				return nil, wantErr
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return cachedJSON, true, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	_, err := svc.GetDailyReport(context.Background(), today(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDailyReport_CacheMiss_Past_QueriesDB(t *testing.T) {
	summaries := makeSummaries(8)

	svc := newService(
		&mockReportRepo{
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		&mockMovementRepo{},
		nil,
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	result, err := svc.GetDailyReport(context.Background(), yesterday(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.TopItems) != 4 {
		t.Errorf("expected 4 items, got %d", len(result.TopItems))
	}
}

func TestGetDailyReport_CacheMiss_Today_RegeneratesSummary(t *testing.T) {
	summaries := makeSummaries(5)
	generateCalled := false

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error {
				generateCalled = true
				return nil
			},
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		&mockMovementRepo{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		&mockItemRepo{
			listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) {
				return nil, nil
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	_, err := svc.GetDailyReport(context.Background(), today(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !generateCalled {
		t.Error("expected GenerateDailySummary to be called for today cache miss")
	}
}

func TestGetDailyReport_CacheMiss_Today_GenerateError(t *testing.T) {
	wantErr := errors.New("aggregate error")

	svc := newService(
		&mockReportRepo{},
		&mockMovementRepo{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return nil, wantErr
			},
		},
		nil,
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	_, err := svc.GetDailyReport(context.Background(), today(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDailyReport_LimitExceedsResults_ClampsToLen(t *testing.T) {
	summaries := makeSummaries(3)

	svc := newService(
		&mockReportRepo{
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		&mockMovementRepo{},
		nil,
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	// limit=10 nhưng chỉ có 3 → trả về 3, không panic
	result, err := svc.GetDailyReport(context.Background(), yesterday(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.TopItems) != 3 {
		t.Errorf("expected 3 items (clamped), got %d", len(result.TopItems))
	}
}

func TestGetDailyReport_DBError(t *testing.T) {
	wantErr := errors.New("db connection lost")

	svc := newService(
		&mockReportRepo{
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				return nil, wantErr
			},
		},
		&mockMovementRepo{},
		nil,
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return "", false, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	_, err := svc.GetDailyReport(context.Background(), yesterday(), 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
