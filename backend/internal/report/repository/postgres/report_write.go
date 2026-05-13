package postgres

import (
	"context"
	"inventory-movement-processing/internal/report/entity"

	"gorm.io/gorm/clause"
)

func (repo *reportRepository) UpsertDailyItemSummary(
	ctx context.Context,
	data []*entity.DailyItemSummary,
) error {

	if len(data) == 0 {
		return nil
	}

	return repo.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "item_id"},
				{Name: "summary_date"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"total_in",
				"total_out",
				"total_adjust",
				"updated_at",
			}),
		}).
		Create(&data).Error
}
