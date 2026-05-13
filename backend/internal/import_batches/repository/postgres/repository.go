package postgres

import (
	"context"
	batchEntity "inventory-movement-processing/internal/import_batches/entity"

	"gorm.io/gorm"
)

type BatchRepository interface {
	Create(ctx context.Context, batch *batchEntity.ImportBatch) error
	GetByID(ctx context.Context, id int32) (*batchEntity.ImportBatch, error)
	Update(ctx context.Context, batch *batchEntity.ImportBatch) error
	GetByFileName(ctx context.Context, fileName string) (*batchEntity.ImportBatch, error)
}

type batchRepository struct {
	db *gorm.DB
}

func NewBatchRepository(db *gorm.DB) BatchRepository {
	return &batchRepository{db: db}
}
