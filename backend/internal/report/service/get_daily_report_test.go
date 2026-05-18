package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
)

func TestIsDataStale(t *testing.T) {
	d := today()
	start := startOfDay(d)

	tests := []struct {
		name      string
		items     []*reportEntity.DailyItemSummary
		wantStale bool
	}{
		{"Empty slice", []*reportEntity.DailyItemSummary{}, true},
		{"Nil UpdatedAt", []*reportEntity.DailyItemSummary{{}}, true},
		{"UpdatedAt before startOfDay (stale)", makeSummariesWithUpdatedAt(1, start.Add(-time.Hour)), true},
		{"UpdatedAt exactly startOfDay (fresh)", makeSummariesWithUpdatedAt(1, start), false},
		{"UpdatedAt after startOfDay (fresh)", makeSummariesWithUpdatedAt(1, start.Add(time.Hour)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDataStale(tt.items, d); got != tt.wantStale {
				t.Errorf("isDataStale() = %v, want %v", got, tt.wantStale)
			}
		})
	}
}

// Edge Cases & Complex Flows
func TestGetDailyReport_LimitExceedsResults_ClampsToLen(t *testing.T) {
	updatedAt := startOfDay(yesterday()).Add(1 * time.Hour)
	summaries := makeSummariesWithUpdatedAt(3, updatedAt)

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

	result, err := svc.GetDailyReport(context.Background(), yesterday(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.TopItems) != 3 {
		t.Errorf("expected 3 items (clamped), got %d", len(result.TopItems))
	}
}

func TestGetDailyReport_CacheHit_Today_Stale_Regenerates(t *testing.T) {
	staleSummaries := makeSummaries(2)
	updatedAt := startOfDay(today()).Add(1 * time.Hour)
	freshSummaries := makeSummariesWithUpdatedAt(5, updatedAt)

	staleJSON := makeCachedJSON(t, staleSummaries, time.Now().Add(-2*time.Minute))

	regenerateCalled := false

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error {
				regenerateCalled = true
				return nil
			},
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				return freshSummaries, nil
			},
		},
		&mockMovementRepo{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return freshSummaries, nil
			},
		},
		&mockItemRepo{
			listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) {
				return nil, nil
			},
		},
		&mockCache{
			getFn: func(_ context.Context, _ string) (string, bool, error) {
				return staleJSON, true, nil
			},
		},
		&mockConfig{cacheLimit: 10},
	)

	result, err := svc.GetDailyReport(context.Background(), today(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !regenerateCalled {
		t.Error("expected regenerate to be called on stale cache")
	}
	if len(result.TopItems) != 5 {
		t.Errorf("expected 5 fresh items, got %d", len(result.TopItems))
	}
}

func TestGetDailyReport_DBStale_ByUpdatedAt_FallbackRegenerate(t *testing.T) {
	staleUpdatedAt := startOfDay(yesterday()).Add(-11 * time.Hour)
	staleSummaries := makeSummariesWithUpdatedAt(3, staleUpdatedAt)

	freshUpdatedAt := startOfDay(yesterday()).Add(1 * time.Minute)
	freshSummaries := makeSummariesWithUpdatedAt(4, freshUpdatedAt)

	listCallCount := 0

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error {
				return nil
			},
			listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
				listCallCount++
				if listCallCount == 1 {
					return staleSummaries, nil
				}
				return freshSummaries, nil
			},
		},
		&mockMovementRepo{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return freshSummaries, nil
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

	result, err := svc.GetDailyReport(context.Background(), yesterday(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listCallCount != 2 {
		t.Errorf("expected 2 DB calls, got %d", listCallCount)
	}
	if len(result.TopItems) != 4 {
		t.Errorf("expected 4 fresh items after fallback, got %d", len(result.TopItems))
	}
}

func TestGetDailyReport_Errors(t *testing.T) {
	tests := []struct {
		name       string
		isToday    bool
		setupMocks func() *reportService
	}{
		{
			name:    "DB list error",
			isToday: false,
			setupMocks: func() *reportService {
				return newService(
					&mockReportRepo{
						listTopActiveFn: func(_ context.Context, _ time.Time, _ int) ([]*reportEntity.DailyItemSummary, error) {
							return nil, errors.New("db error")
						},
					},
					&mockMovementRepo{}, nil, &mockCache{
						getFn: func(_ context.Context, _ string) (string, bool, error) { return "", false, nil },
					}, &mockConfig{cacheLimit: 10},
				)
			},
		},
		{
			name:    "Cache hit today but low stock fetch fails",
			isToday: true,
			setupMocks: func() *reportService {
				summaries := makeSummaries(2)
				cachedJSON := makeCachedJSON(t, summaries, time.Now())
				return newService(
					&mockReportRepo{}, &mockMovementRepo{},
					&mockItemRepo{
						listLowStockFn: func(_ context.Context) ([]*itemEntity.Item, error) {
							return nil, errors.New("low stock error")
						},
					},
					&mockCache{
						getFn: func(_ context.Context, _ string) (string, bool, error) { return cachedJSON, true, nil },
					}, &mockConfig{cacheLimit: 10},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupMocks()
			date := yesterday()
			if tt.isToday {
				date = today()
			}
			_, err := svc.GetDailyReport(context.Background(), date, 5)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func marshalCachedReport(t *testing.T, r reportEntity.CachedReport) string {
	t.Helper()
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal CachedReport: %v", err)
	}
	return string(raw)
}

func makeCachedJSON(t *testing.T, summaries []*reportEntity.DailyItemSummary, generatedAt time.Time) string {
	t.Helper()
	return marshalCachedReport(t, reportEntity.CachedReport{
		Items:       summaries,
		GeneratedAt: generatedAt,
	})
}

func makeSummariesWithUpdatedAt(n int, updatedAt time.Time) []*reportEntity.DailyItemSummary {
	items := make([]*reportEntity.DailyItemSummary, n)
	t := updatedAt
	for i := range items {
		items[i] = &reportEntity.DailyItemSummary{}
		items[i].UpdatedAt = &t
	}
	return items
}

func startOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}
