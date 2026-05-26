package service

import (
	"context"
	"testing"
	"time"

	reportEntity "inventory-movement-processing/internal/report/entity"
)

func BenchmarkGetDailyReport(b *testing.B) {
	// 1. Khởi tạo Mock Repo với độ trễ của Database (~50ms)
	mockRepo := &mockReportRepo{
		listTopActiveFn: func(ctx context.Context, date time.Time, limit int) ([]*reportEntity.DailyItemSummary, error) {
			time.Sleep(50 * time.Millisecond) // Giả lập DB query chậm
			return []*reportEntity.DailyItemSummary{
				{ItemID: 1, SummaryDate: date},
			}, nil
		},
		upsertFn: func(ctx context.Context, data []*reportEntity.DailyItemSummary) error {
			time.Sleep(50 * time.Millisecond) // Giả lập DB write
			return nil
		},
	}

	// 2. Khởi tạo Mock Cache với độ trễ của Redis (~2ms)
	// Trả về false (Cache Miss) để ÉP các request phải đi xuống DB,
	// từ đó kiểm tra xem Singleflight có chặn được Stampede hay không.
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
			return true, nil // Giả lập luôn lấy được Lock thành công
		},
	}

	mockMovement := &mockMovementUseCase{
		aggregateFn: func(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error) {
			time.Sleep(20 * time.Millisecond)
			return nil, nil
		},
	}
	mockItem := &mockItemRepo{} // Không dùng tới trong test này nên để trống
	mockLog := &mockLogger{}

	// 3. Định nghĩa các kịch bản test
	scenarios := []struct {
		name string
		cfg  ReportCacheConfig
	}{
		{
			name: "1_All_Enabled", // Chạy qua Singleflight -> Lock -> Cache/DB
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
			name: "4_Direct_DB_Only", // Bỏ qua tất cả, đâm thẳng vào DB
			cfg: func() ReportCacheConfig {
				c := DefaultReportCacheConfig()
				c.DisableCache = true
				c.DisableSingleFlight = true
				c.DisableLock = true
				return c
			}(),
		},
	}

	// 4. Chạy vòng lặp benchmark
	ctx := context.Background()
	targetDate := time.Now()

	for _, s := range scenarios {
		b.Run(s.name, func(b *testing.B) {
			svc := newService(mockRepo, mockMovement, mockItem, mockCacheStorage, s.cfg, mockLog)

			b.ResetTimer() // Xóa thời gian setup khỏi kết quả đo

			// Tạo hàng ngàn request đồng thời
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					// Gọi hàm cần benchmark
					_, _ = svc.GetDailyReport(ctx, targetDate, 5)
				}
			})
		})
	}
}
