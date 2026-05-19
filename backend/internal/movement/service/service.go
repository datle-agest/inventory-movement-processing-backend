package service

import (
	"context"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/core"
	"mime/multipart"
	"time"
)

type movementRepository interface {
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	Create(ctx context.Context, movement *movementEntity.Movement) error
	AggregateDailyItemSummaryFromMovement(ctx context.Context, start time.Time, end time.Time) ([]*reportEntity.DailyItemSummary, error)
}

type itemRepository interface {
	GetItemForUpdate(ctx context.Context, id int32) (*itemEntity.Item, error)
	UpdateStock(ctx context.Context, itemID int32, quantity int32) error
}

type MovementService interface {
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (movementEntity.ImportBatchResult, error)
	ProcessOne(ctx context.Context, m *movementEntity.Movement) (movementEntity.ProcessStatus, error)
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	AggregateDailyItemSummaryFromMovement(ctx context.Context, date time.Time) ([]*reportEntity.DailyItemSummary, error)
}

type service struct {
	movementRepo movementRepository
	itemRepo     itemRepository   // Inject thêm Repo của Item
	txManager    common.TxManager // Inject thêm TxManager
	workerPool   workerc.WorkerPool
}

func NewMovementService(
	movementRepo movementRepository,
	itemRepo itemRepository,
	txManager common.TxManager,
	workerPool workerc.WorkerPool,
) MovementService {
	return &service{
		movementRepo: movementRepo,
		itemRepo:     itemRepo,
		txManager:    txManager,
		workerPool:   workerPool,
	}
}
