package service

import (
	"context"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/components/workerc"
	"inventory-movement-processing/pkg/core"
	"mime/multipart"
)

type movementRepository interface {
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
	Create(ctx context.Context, movement *movementEntity.Movement) error
	ProcessMovement(ctx context.Context, m *movementEntity.Movement) error
}

type MovementService interface {
	ImportBatch(ctx context.Context, file *multipart.FileHeader) (movementEntity.ImportBatchResult, error)
	ProcessOne(ctx context.Context, m *movementEntity.Movement) (movementEntity.ProcessStatus, error)
	GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*movementEntity.Movement, error)
}

type service struct {
	movementRepo movementRepository
	workerPool   workerc.WorkerPool
}

func NewMovementService(movementRepo movementRepository, workerPool workerc.WorkerPool) MovementService {
	return &service{
		movementRepo: movementRepo,
		workerPool:   workerPool,
	}
}
