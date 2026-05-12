package postgres

import (
	"context"
	"inventory-movement-processing/internal/report/entity"
)

func (repo *reportRepository) CreateReport(ctx context.Context, data *entity.Report) error {
	return repo.db.WithContext(ctx).Create(data).Error
}
