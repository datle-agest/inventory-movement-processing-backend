package service

import (
	"context"
	"encoding/json"
	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"sync"
	"sync/atomic"
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

	cacheCfg := ReportCacheConfig{}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, mockLog)

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

	cacheCfg := ReportCacheConfig{}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, mockLog)

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

	cacheCfg := ReportCacheConfig{}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, mockLog)

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

	cacheCfg := ReportCacheConfig{}
	mockLog := &mockLogger{}

	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, mockLog)

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

// =========================================================================
// TEST CASE P1: Cache stampede — distributed lock
// =========================================================================

// TestGetDailyReport_LockHolder_GeneratesOnce verifies that the goroutine
// that acquires the lock calls GenerateDailySummary exactly once.
func TestGetDailyReport_LockHolder_GeneratesOnce(t *testing.T) {
	requestedDate := time.Now()

	dbItems := []*reportEntity.DailyItemSummary{
		{ItemID: 1, SummaryDate: requestedDate, TotalIn: 10},
	}

	var generateCalled int32

	mockRepo := &mockReportRepo{
		upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error { return nil },
		listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
			return dbItems, nil
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
			atomic.AddInt32(&generateCalled, 1)
			return dbItems, nil
		},
	}

	mockItem := &mockItemRepo{
		listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) { return nil, nil },
	}

	mockCacheStore := &mockCache{
		getFn: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, nil // always cache miss
		},
		setNXFn: func(_ context.Context, _ string, _ string, _ time.Duration) (bool, error) {
			return true, nil // always acquire lock
		},
	}

	cacheCfg := DefaultReportCacheConfig()
	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, &mockLogger{})

	_, err := service.GetDailyReport(context.Background(), requestedDate, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&generateCalled); got != 1 {
		t.Errorf("expected GenerateDailySummary called 1 time, got %d", got)
	}
}

// TestGetDailyReport_LockContender_SkipsGenerate verifies that when the lock
// is already held, the goroutine queries DB directly without calling GenerateDailySummary.
func TestGetDailyReport_LockContender_SkipsGenerate(t *testing.T) {
	requestedDate := time.Now()

	dbItems := []*reportEntity.DailyItemSummary{
		{ItemID: 2, SummaryDate: requestedDate, TotalIn: 20},
	}

	var generateCalled int32

	mockRepo := &mockReportRepo{
		upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error { return nil },
		listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
			return dbItems, nil
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
			atomic.AddInt32(&generateCalled, 1)
			return dbItems, nil
		},
	}

	mockItem := &mockItemRepo{
		listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) { return nil, nil },
	}

	mockCacheStore := &mockCache{
		getFn: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, nil
		},
		setNXFn: func(_ context.Context, _ string, _ string, _ time.Duration) (bool, error) {
			return false, nil // lock NOT acquired — another instance holds it
		},
	}

	cacheCfg := DefaultReportCacheConfig()
	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, &mockLogger{})

	result, err := service.GetDailyReport(context.Background(), requestedDate, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&generateCalled); got != 0 {
		t.Errorf("lock contender must NOT call GenerateDailySummary, got %d call(s)", got)
	}

	if len(result.TopItems) == 0 {
		t.Error("expected items from DB even without lock")
	}
}

// TestGetDailyReport_Concurrent_OnlyOnceGenerates simulates N goroutines hitting
// GetDailyReport simultaneously. Only the one that wins the lock should call
// GenerateDailySummary; the rest serve from DB directly.
func TestGetDailyReport_Concurrent_OnlyOnceGenerates(t *testing.T) {
	requestedDate := time.Now()

	dbItems := []*reportEntity.DailyItemSummary{
		{ItemID: 3, SummaryDate: requestedDate, TotalIn: 30},
	}

	var generateCalled int32

	// Simulate atomic Redis SetNX: only the first caller gets true
	var lockHeld int32

	mockRepo := &mockReportRepo{
		upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error { return nil },
		listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
			return dbItems, nil
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
			atomic.AddInt32(&generateCalled, 1)
			time.Sleep(50 * time.Millisecond)
			return dbItems, nil
		},
	}

	mockItem := &mockItemRepo{
		listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) { return nil, nil },
	}

	mockCacheStore := &mockCache{
		getFn: func(_ context.Context, _ string) (string, bool, error) {
			return "", false, nil
		},
		setNXFn: func(_ context.Context, _ string, _ string, _ time.Duration) (bool, error) {
			acquired := atomic.CompareAndSwapInt32(&lockHeld, 0, 1)
			return acquired, nil
		},
		delFn: func(_ context.Context, _ ...string) (int64, error) {
			atomic.StoreInt32(&lockHeld, 0)
			return 1, nil
		},
	}

	cacheCfg := DefaultReportCacheConfig()
	service := newService(mockRepo, mockMovement, mockItem, mockCacheStore, cacheCfg, &mockLogger{})

	const concurrency = 20
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			_, _ = service.GetDailyReport(context.Background(), requestedDate, 3)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&generateCalled); got > 1 {
		t.Errorf("expected GenerateDailySummary called at most 1 time across %d goroutines, got %d", concurrency, got)
	}
}
