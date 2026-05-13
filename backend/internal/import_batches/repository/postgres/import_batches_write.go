package postgres

import (
	"context"
	batchEntity "inventory-movement-processing/internal/import_batches/entity"
)

// Create - Tạo batch mới
func (r *batchRepository) Create(ctx context.Context, batch *batchEntity.ImportBatch) error {
	if err := r.db.WithContext(ctx).Create(batch).Error; err != nil {
		return err
	}
	return nil
}

// Update - Cập nhật batch
func (r *batchRepository) Update(ctx context.Context, batch *batchEntity.ImportBatch) error {
	if err := r.db.WithContext(ctx).Save(batch).Error; err != nil {
		return err
	}
	return nil
}
