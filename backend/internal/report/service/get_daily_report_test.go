package service

import (
	"context"
	"encoding/json"
	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"testing"
	"time"
)

// =========================================================================
// TEST CASE 1: Yesterday data is stale -> should trigger regenerate
// =========================================================================
func TestGetDailyReport_YesterdayDataIsStale_ShouldTriggerRegenerate(t *testing.T) {
	requestedDate := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)

	// DB data was only updated until 14:00 on the requested day
	dbUpdatedAt := time.Date(2026, time.May, 18, 14, 0, 0, 0, time.UTC)

	staleItems := []*reportEntity.DailyItemSummary{
		{ItemID: 1, SummaryDate: requestedDate},
	}
	staleItems[0].UpdatedAt = &dbUpdatedAt

	generateSummaryCalled := false

	mockRepo := &mockReportRepo{
		listTopActiveFn: func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
			// First call: return stale data
			if !generateSummaryCalled {
				return staleItems, nil
			}

			// Second call after regenerate: return refreshed data
			freshTime := time.Date(2026, time.May, 19, 1, 0, 0, 0, time.UTC)
			staleItems[0].UpdatedAt = &freshTime

			return staleItems, nil
		},
		upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
			return nil
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
			generateSummaryCalled = true
			return staleItems, nil
		},
	}

	mockItem := &mockItemRepo{}

	mockCacheStore := &mockCache{
		getFn: func(ctx context.Context, key string) (string, bool, error) {
			return "", false, nil
		},
		setFn: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return nil
		},
	}

	mockCfg := &mockConfig{cacheLimit: 10}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, mockCfg, mockLog)

	// RUN
	_, err := service.GetDailyReport(context.Background(), requestedDate, 5)

	// VERIFY
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !generateSummaryCalled {
		t.Errorf("expected stale yesterday data to trigger regenerate summary")
	}
}

// =========================================================================
// TEST CASE 2: Cache hit and data is still fresh
// =========================================================================
func TestGetDailyReport_CacheHit_Fresh(t *testing.T) {
	requestedDate := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)

	cachedItems := []*reportEntity.DailyItemSummary{
		{ItemID: 11, TotalIn: 50, SummaryDate: requestedDate},
	}

	cachedReport := reportEntity.CachedReport{
		Items:       cachedItems,
		GeneratedAt: time.Now().Add(-30 * time.Second),
	}

	rawJSON, _ := json.Marshal(cachedReport)

	mockCacheStore := &mockCache{
		getFn: func(ctx context.Context, key string) (string, bool, error) {
			return string(rawJSON), true, nil
		},
	}

	dbCalled := false

	mockRepo := &mockReportRepo{
		listTopActiveFn: func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
			dbCalled = true
			return nil, nil
		},
	}

	mockMovement := &mockMovementUseCase{}
	mockItem := &mockItemRepo{}
	mockCfg := &mockConfig{cacheLimit: 10}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, mockCfg, mockLog)

	// RUN
	result, err := service.GetDailyReport(context.Background(), requestedDate, 5)

	// VERIFY
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dbCalled {
		t.Errorf("expected fresh cache data to prevent DB query")
	}

	if len(result.TopItems) != 1 || result.TopItems[0].ItemID != 11 {
		t.Errorf("returned data does not match cached data")
	}
}

// =========================================================================
// TEST CASE 3: Cache hit but today's cache is stale
// =========================================================================
func TestGetDailyReport_CacheHit_StaleToday(t *testing.T) {
	today := time.Now()

	cachedItems := []*reportEntity.DailyItemSummary{
		{ItemID: 22, SummaryDate: today},
	}

	cachedReport := reportEntity.CachedReport{
		Items:       cachedItems,
		GeneratedAt: today.Add(-5 * time.Minute),
	}

	rawJSON, _ := json.Marshal(cachedReport)

	mockCacheStore := &mockCache{
		getFn: func(ctx context.Context, key string) (string, bool, error) {
			return string(rawJSON), true, nil
		},
		setFn: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return nil
		},
	}

	dbCalled := false

	mockRepo := &mockReportRepo{
		listTopActiveFn: func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
			dbCalled = true

			freshTime := time.Now()
			cachedItems[0].UpdatedAt = &freshTime

			return cachedItems, nil
		},
		upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
			return nil
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
			return cachedItems, nil
		},
	}

	mockItem := &mockItemRepo{
		listLowStockFn: func(ctx context.Context) ([]*itemEntity.Item, error) {
			return nil, nil
		},
	}

	mockCfg := &mockConfig{cacheLimit: 10}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, mockCfg, mockLog)

	// RUN
	_, err := service.GetDailyReport(context.Background(), today, 5)

	// VERIFY
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !dbCalled {
		t.Errorf("expected stale cache for today to trigger DB query")
	}
}

// =========================================================================
// TEST CASE 4: Empty DB should trigger fallback regenerate
// =========================================================================
func TestGetDailyReport_PastDate_EmptyDB_TriggersFallback(t *testing.T) {
	requestedDate := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)

	mockCacheStore := &mockCache{
		getFn: func(ctx context.Context, key string) (string, bool, error) {
			return "", false, nil
		},
		setFn: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return nil
		},
	}

	generateSummaryCalled := false

	mockRepo := &mockReportRepo{
		listTopActiveFn: func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
			// First call: DB is empty
			if !generateSummaryCalled {
				return []*reportEntity.DailyItemSummary{}, nil
			}

			// Second call after fallback regenerate
			freshUpdatedAt := time.Date(2026, time.May, 19, 4, 0, 0, 0, time.UTC)

			items := []*reportEntity.DailyItemSummary{
				{ItemID: 55, SummaryDate: requestedDate},
			}

			items[0].UpdatedAt = &freshUpdatedAt

			return items, nil
		},
		upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
			return nil
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
			generateSummaryCalled = true
			return []*reportEntity.DailyItemSummary{}, nil
		},
	}

	mockItem := &mockItemRepo{}
	mockCfg := &mockConfig{cacheLimit: 10}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, mockCfg, mockLog)

	// RUN
	result, err := service.GetDailyReport(context.Background(), requestedDate, 5)

	// VERIFY
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !generateSummaryCalled {
		t.Errorf("expected empty DB result to trigger fallback regenerate")
	}

	if result == nil || len(result.TopItems) != 1 || result.TopItems[0].ItemID != 55 {
		t.Errorf("unexpected fallback result data")
	}
}
