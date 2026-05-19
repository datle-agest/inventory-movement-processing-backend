package service

import (
	"context"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/core"
	"time"
)

// --- movementRepository mock ---
type mockMovementRepo struct {
	getMovementsByItemIDFn                  func(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	createFn                                func(ctx context.Context, movement *movementEntity.Movement) error
	processMovementFn                       func(ctx context.Context, m *movementEntity.Movement) error
	aggregateDailyItemSummaryFromMovementFn func(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error)
}

func (m *mockMovementRepo) GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error) {
	if m.getMovementsByItemIDFn != nil {
		return m.getMovementsByItemIDFn(ctx, itemId, paging)
	}
	return nil, nil
}

func (m *mockMovementRepo) Create(ctx context.Context, movement *movementEntity.Movement) error {
	if m.createFn != nil {
		return m.createFn(ctx, movement)
	}
	return nil
}

func (m *mockMovementRepo) ProcessMovement(ctx context.Context, movement *movementEntity.Movement) error {
	if m.processMovementFn != nil {
		return m.processMovementFn(ctx, movement)
	}
	return nil
}

func (m *mockMovementRepo) AggregateDailyItemSummaryFromMovement(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error) {
	if m.aggregateDailyItemSummaryFromMovementFn != nil {
		return m.aggregateDailyItemSummaryFromMovementFn(ctx, start, end)
	}
	return nil, nil
}

// --- workerPool mock ---
type mockWorkerPool struct {
	submitFn func(job workerc.Job)
	waitFn   func()
}

func (m *mockWorkerPool) Submit(job workerc.Job) {
	if m.submitFn != nil {
		m.submitFn(job)
		return
	}
	// Synchronous execution for deterministic unit tests
	job()
}

func (m *mockWorkerPool) Wait() {
	if m.waitFn != nil {
		m.waitFn()
	}
}

// Helpers
func newMovementService(
	mr movementRepository,
	wp workerc.WorkerPool,
) MovementService {
	return NewMovementService(mr, wp)
}
