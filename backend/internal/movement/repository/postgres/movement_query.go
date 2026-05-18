package postgres

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/core"
	"time"
)

func (r *movementRepository) AggregateDailyItemSummaryFromMovement(
	ctx context.Context,
	date time.Time,
) ([]*reportEntity.DailyItemSummary, error) {

	var results []*reportEntity.DailyItemSummary
	var tmp entity.Movement

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Table(tmp.TableName()).
		Select(`
			item_id,
			DATE(created_at) as summary_date,
			SUM(CASE WHEN movement_type = 'IN' THEN quantity ELSE 0 END) as total_in,
			SUM(CASE WHEN movement_type = 'OUT' THEN quantity ELSE 0 END) as total_out,
			SUM(CASE WHEN movement_type = 'ADJUST' THEN ABS(quantity) ELSE 0 END) as total_adjust
		`).
		Where("created_at >= ? AND created_at < ?", start, end).
		Group("item_id, DATE(created_at)").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *movementRepository) GetMovementsByItemID(ctx context.Context, itemId int, paging *core.Pagination) ([]*entity.Movement, error) {
	var movements []*entity.Movement
	query := r.db.WithContext(ctx).Model(&entity.Movement{}).Where("item_id = ?", itemId)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	paging.Total = int(total)

	offset := (paging.Page - 1) * paging.Limit
	err := query.Order("id DESC").Offset(offset).Limit(paging.Limit).Find(&movements).Error

	return movements, err
}
