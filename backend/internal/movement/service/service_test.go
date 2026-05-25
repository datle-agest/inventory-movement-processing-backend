package service

import (
	"context"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/core"
	"inventory-movement-processing/pkg/logger"
	"time"
)

// --- mockMovementRepo ---
type mockMovementRepo struct {
	getMovementsByItemIDFn                  func(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	createFn                                func(ctx context.Context, movement *movementEntity.Movement) error
	aggregateDailyItemSummaryFromMovementFn func(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error)
}

func (m *mockMovementRepo) GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error) {
	return m.getMovementsByItemIDFn(ctx, itemId, paging)
}
func (m *mockMovementRepo) Create(ctx context.Context, movement *movementEntity.Movement) error {
	return m.createFn(ctx, movement)
}
func (m *mockMovementRepo) AggregateDailyItemSummaryFromMovement(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error) {
	return m.aggregateDailyItemSummaryFromMovementFn(ctx, start, end)
}

// --- mockItemService ---
type mockItemService struct {
	adjustStockFn func(ctx context.Context, itemID int32, quantityChange int32) error
}

func (m *mockItemService) AdjustStock(ctx context.Context, itemID int32, quantityChange int32) error {
	return m.adjustStockFn(ctx, itemID, quantityChange)
}

// --- mockTxManager — implements common.TxManager ---
type mockTxManager struct {
	withTxFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

// --- mockWorkerPool ---
type mockWorkerPool struct {
	submitFn func(job workerc.Job)
}

func (m *mockWorkerPool) Submit(job workerc.Job) {
	if m.submitFn != nil {
		m.submitFn(job)
		return
	}
	job()
}

// --- mockLogger ---
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{}) {}
func (m *mockLogger) Info(args ...interface{})  {}
func (m *mockLogger) Warn(args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{}) {}

func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}

func (m *mockLogger) With(key string, value interface{}) logger.Logger { return m }
func (m *mockLogger) WithFields(fields logger.Fields) logger.Logger    { return m }

// --- helper ---
func newMovementService(
	repo movementRepository,
	itemSvc itemService,
	tx *mockTxManager,
	wp workerc.WorkerPool,
	log logger.Logger,
) MovementService {
	return NewMovementService(repo, itemSvc, tx, wp, log)
}
