package service

import (
	"context"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/logger"
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

// --- movementUseCase mock ---
type mockMovementUseCase struct {
	aggregateFn func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error)
}

func (m *mockMovementUseCase) AggregateDailyItemSummaryFromMovement(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
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

// --- Logger mock ---
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{}) {}
func (m *mockLogger) Info(args ...interface{})  {}
func (m *mockLogger) Warn(args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{}) {}

func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}

func (m *mockLogger) With(key string, value interface{}) logger.Logger {
	return m
}

func (m *mockLogger) WithFields(fields logger.Fields) logger.Logger {
	return m
}

// Helpers

func newService(
	rr reportRepository,
	mu movementService,
	ir itemService,
	cache *mockCache,
	cfg *mockConfig,
	log logger.Logger,
) *reportService {
	return NewReportService(rr, mu, ir, cache, cfg, log)
}
