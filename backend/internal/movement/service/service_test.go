package service

import (
	"context"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/core"
	"time"
)

// --- itemRepository mock ---
type mockItemRepo struct {
	getItemForUpdateFn func(ctx context.Context, id int32) (*itemEntity.Item, error)
	updateStockFn      func(ctx context.Context, itemID int32, quantity int32) error
}

func (m *mockItemRepo) GetItemForUpdate(ctx context.Context, id int32) (*itemEntity.Item, error) {
	if m.getItemForUpdateFn != nil {
		return m.getItemForUpdateFn(ctx, id)
	}
	return nil, nil
}

func (m *mockItemRepo) UpdateStock(ctx context.Context, itemID int32, quantity int32) error {
	if m.updateStockFn != nil {
		return m.updateStockFn(ctx, itemID, quantity)
	}
	return nil
}

// --- TxManager mock ---
type mockTxManager struct {
	withTxFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx) // Unit test mặc định chạy đồng bộ thẳng luôn
}

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

// Helpers
func newMovementService(
	mr movementRepository,
	wp workerc.WorkerPool,
) MovementService {
	// Tự động inject mock mặc định vào
	return NewMovementService(mr, &mockItemRepo{}, &mockTxManager{}, wp)
}
