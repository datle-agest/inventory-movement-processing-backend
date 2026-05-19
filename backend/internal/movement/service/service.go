package service

import (
	"context"
	"inventory-movement-processing/common"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/core"
	"inventory-movement-processing/pkg/logger"
	"mime/multipart"
	"time"
)

type movementRepository interface {
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	Create(ctx context.Context, movement *movementEntity.Movement) error
	AggregateDailyItemSummaryFromMovement(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error)
}

type itemService interface {
	AdjustStock(ctx context.Context, itemID int32, quantityChange int32) error
}

type MovementService interface {
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (movementEntity.ImportBatchResult, error)
	ProcessOne(ctx context.Context, m *movementEntity.Movement) (movementEntity.ProcessStatus, error)
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	AggregateDailyItemSummaryFromMovement(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error)
}

type service struct {
	movementRepo movementRepository
	itemService  itemService
	txManager    common.TxManager
	workerPool   workerc.WorkerPool
	logger       logger.Logger
}

func NewMovementService(
	movementRepo movementRepository,
	itemService itemService,
	txManager common.TxManager,
	workerPool workerc.WorkerPool,
	logger logger.Logger,
) MovementService {
	return &service{
		movementRepo: movementRepo,
		itemService:  itemService,
		txManager:    txManager,
		workerPool:   workerPool,
		logger:       logger,
	}
}
