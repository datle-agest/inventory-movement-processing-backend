package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
)

// --- reportRepository mock ---
type mockReportRepo struct {
	upsertFn        func(ctx context.Context, data []*reportEntity.DailyItemSummary) error
	listTopActiveFn func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error)
}

func (m *mockReportRepo) UpsertDailyItemSummary(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
	return m.upsertFn(ctx, data)
}

func (m *mockReportRepo) ListTopActiveItemsByDate(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
	return m.listTopActiveFn(ctx, date, limit)
}

// --- movementRepository mock ---
type mockMovementRepo struct {
	aggregateFn func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error)
}

func (m *mockMovementRepo) AggregateDailyItemSummaryFromMovement(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
	return m.aggregateFn(ctx, date)
}

// --- itemRepository mock ---
type mockItemRepo struct {
	listLowStockFn func(ctx context.Context) ([]*itemEntity.Item, error)
}

func (m *mockItemRepo) ListLowStockItems(ctx context.Context) ([]*itemEntity.Item, error) {
	return m.listLowStockFn(ctx)
}

// --- CacheProvider mock ---
type mockCache struct {
	getFn func(ctx context.Context, key string) (string, bool, error)
	setFn func(ctx context.Context, key string, value string, ttl time.Duration) error
	delFn func(ctx context.Context, keys ...string) (int64, error)
}

func (m *mockCache) Get(ctx context.Context, key string) (string, bool, error) {
	return m.getFn(ctx, key)
}

func (m *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCache) Del(ctx context.Context, keys ...string) (int64, error) {
	if m.delFn != nil {
		return m.delFn(ctx, keys...)
	}
	return 0, nil
}

// --- Config mock ---
type mockConfig struct {
	cacheLimit int
}

func (m *mockConfig) GetReportCacheLimit() int { return m.cacheLimit }

// Helpers

func newService(
	rr reportRepository,
	mr movementRepository,
	ir itemRepository,
	cache *mockCache,
	cfg *mockConfig,
) *reportService {
	return NewReportService(rr, mr, ir, cache, cfg)
}

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func yesterday() time.Time {
	return today().AddDate(0, 0, -1)
}

func makeSummaries(n int) []*reportEntity.DailyItemSummary {
	items := make([]*reportEntity.DailyItemSummary, n)
	for i := range items {
		items[i] = &reportEntity.DailyItemSummary{}
	}
	return items
}

func marshalSummaries(t *testing.T, summaries []*reportEntity.DailyItemSummary) string {
	t.Helper()
	raw, err := json.Marshal(summaries)
	if err != nil {
		t.Fatalf("marshal summaries: %v", err)
	}
	return string(raw)
}
