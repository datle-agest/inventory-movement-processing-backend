package postgres

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"time"
)

func (repo *movementRepository) AggregateDailyItemSummaryFromMovement(
	ctx context.Context,
	date time.Time,
) ([]*reportEntity.DailyItemSummary, error) {

	var results []*reportEntity.DailyItemSummary
	var tmp entity.Movement

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	err := repo.db.WithContext(ctx).
		Table(tmp.TableName()).
		Select(`
			item_id,
			DATE(created_at) as summary_date,
			SUM(CASE WHEN movement_type = 'IN' THEN quantity ELSE 0 END) as total_in,
			SUM(CASE WHEN movement_type = 'OUT' THEN quantity ELSE 0 END) as total_out,
			SUM(CASE WHEN movement_type = 'ADJUST' THEN quantity ELSE 0 END) as total_adjust
		`).
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("item_id, DATE(created_at)").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
