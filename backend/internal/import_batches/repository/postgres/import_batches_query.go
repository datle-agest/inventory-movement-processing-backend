package postgres

import (
	"context"
	batchEntity "inventory-movement-processing/internal/import_batches/entity"
)

// GetByFileName - Lấy batch theo tên file
func (r *batchRepository) GetByFileName(ctx context.Context, fileName string) (*batchEntity.ImportBatch, error) {
	var batch batchEntity.ImportBatch
	if err := r.db.WithContext(ctx).Where("file_name = ?", fileName).First(&batch).Error; err != nil {
		return nil, err
	}
	return &batch, nil
}

// GetByID - Lấy batch theo ID
func (r *batchRepository) GetByID(ctx context.Context, id int32) (*batchEntity.ImportBatch, error) {
	var batch batchEntity.ImportBatch
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&batch).Error; err != nil {
		return nil, err
	}
	return &batch, nil
}
