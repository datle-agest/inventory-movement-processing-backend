package service

import (
	"context"
	"errors"
	"testing"
	"time"

	reportEntity "inventory-movement-processing/internal/report/entity"
)

func TestGenerateDailySummary_Success(t *testing.T) {
	summaries := makeSummaries(3)

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, data []*reportEntity.DailyItemSummary) error {
				if len(data) != 3 {
					t.Errorf("expected 3 summaries, got %d", len(data))
				}
				return nil
			},
		},
		&mockMovementUseCase{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return summaries, nil
			},
		},
		nil,
		&mockCache{getFn: nil},
		&mockConfig{cacheLimit: 10},
	)

	err := svc.GenerateDailySummary(context.Background(), today())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateDailySummary_EmptySummaries_SkipsUpsert(t *testing.T) {
	upsertCalled := false

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error {
				upsertCalled = true
				return nil
			},
		},
		&mockMovementUseCase{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return []*reportEntity.DailyItemSummary{}, nil
			},
		},
		nil,
		&mockCache{getFn: nil},
		&mockConfig{},
	)

	err := svc.GenerateDailySummary(context.Background(), today())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if upsertCalled {
		t.Error("upsert should not be called when summaries is empty")
	}
}

func TestGenerateDailySummary_AggregateError(t *testing.T) {
	wantErr := errors.New("aggregate failed")

	svc := newService(
		&mockReportRepo{},
		&mockMovementUseCase{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return nil, wantErr
			},
		},
		nil,
		&mockCache{getFn: nil},
		&mockConfig{},
	)

	err := svc.GenerateDailySummary(context.Background(), today())
	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}

func TestGenerateDailySummary_UpsertError(t *testing.T) {
	wantErr := errors.New("upsert failed")

	svc := newService(
		&mockReportRepo{
			upsertFn: func(_ context.Context, _ []*reportEntity.DailyItemSummary) error {
				return wantErr
			},
		},
		&mockMovementUseCase{
			aggregateFn: func(_ context.Context, _ time.Time) ([]*reportEntity.DailyItemSummary, error) {
				return makeSummaries(2), nil
			},
		},
		nil,
		&mockCache{getFn: nil},
		&mockConfig{},
	)

	err := svc.GenerateDailySummary(context.Background(), today())
	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}
