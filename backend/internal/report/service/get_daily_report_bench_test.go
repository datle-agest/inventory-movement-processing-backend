package service

import (
	"context"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
)

// go test -bench=BenchmarkGetDailyReport -benchmem -cpu=1,4,8,16
func BenchmarkGetDailyReport(b *testing.B) {
	dbConnectionPool := make(chan struct{}, 10)

	// 1. Initialize Mock Repo with Database-like latency (~50ms)
	mockRepo := &mockReportRepo{
		listTopActiveFn: func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
			// Queue for a DB connection. If the pool is full,
			// the goroutine will be blocked by the Go scheduler.
			dbConnectionPool <- struct{}{}

			// Always release the connection back to the pool after query completion
			defer func() { <-dbConnectionPool }()

			// The request now actually occupies the DB connection and starts querying
			time.Sleep(50 * time.Millisecond) // Simulate I/O delay

			return []*reportEntity.DailyItemSummary{
				{ItemID: 1, SummaryDate: date},
			}, nil
		},
		upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
			dbConnectionPool <- struct{}{}
			defer func() { <-dbConnectionPool }()

			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}

	// 2. Initialize Mock Cache with Redis-like latency (~2ms)
	// Always return false (Cache Miss) to FORCE requests
	// to hit the DB, allowing us to verify whether
	// Singleflight can prevent cache stampede.
	mockCacheStorage := &mockCache{
		getFn: func(ctx context.Context, key string) (string, bool, error) {
			time.Sleep(2 * time.Millisecond)
			return "", false, nil
		},
		setFn: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			time.Sleep(2 * time.Millisecond)
			return nil
		},
		setNXFn: func(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
			time.Sleep(2 * time.Millisecond)
			return true, nil // Simulate successful lock acquisition every time
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
			time.Sleep(20 * time.Millisecond)
			return nil, nil
		},
	}

	mockItem := &mockItemRepo{
		listLowStockFn: func(ctx context.Context) ([]*itemEntity.Item, error) {
			time.Sleep(5 * time.Millisecond)
			return nil, nil
		},
	}

	mockLog := &mockLogger{}

	// 3. Define benchmark scenarios
	scenarios := []struct {
		name string
		cfg  ReportCacheConfig
	}{
		{
			name: "1_All_Enabled", // Flow: Singleflight -> Lock -> Cache/DB
			cfg: func() ReportCacheConfig {
				c := DefaultReportCacheConfig()
				return c
			}(),
		},
		{
			name: "2_Disable_SingleFlight",
			cfg: func() ReportCacheConfig {
				c := DefaultReportCacheConfig()
				c.DisableSingleFlight = true
				return c
			}(),
		},
		{
			name: "3_Disable_SF_And_Lock",
			cfg: func() ReportCacheConfig {
				c := DefaultReportCacheConfig()
				c.DisableSingleFlight = true
				c.DisableLock = true
				return c
			}(),
		},
		{
			name: "4_Direct_DB_Only", // Skip everything and hit the DB directly
			cfg: func() ReportCacheConfig {
				c := DefaultReportCacheConfig()
				c.DisableCache = true
				c.DisableSingleFlight = true
				c.DisableLock = true
				return c
			}(),
		},
	}

	// 4. Run benchmark loop
	ctx := context.Background()
	targetDate := time.Now()

	for _, s := range scenarios {
		b.Run(s.name, func(b *testing.B) {
			svc := newService(mockRepo, mockMovement, mockItem, mockCacheStorage, s.cfg, mockLog)

			b.ResetTimer() // Exclude setup time from benchmark result

			// Generate thousands of concurrent requests
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					// Call the function under benchmark
					_, _ = svc.GetDailyReport(ctx, targetDate, 5)
				}
			})
		})
	}
}
