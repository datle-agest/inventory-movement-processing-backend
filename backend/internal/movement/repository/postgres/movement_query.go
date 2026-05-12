package postgres

import (
	"context"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	"time"
)

func (r *movementRepository) GetByDateRange(ctx context.Context, from, to time.Time) ([]movementEntity.Movement, error) {
	var movements []movementEntity.Movement

	err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", from, to).
		Find(&movements).Error
	if err != nil {
		return nil, err
	}

	return movements, nil
}

func (r *movementRepository) GetAggregatedByItem(ctx context.Context, from, to time.Time) (map[int32]int32, error) {
	type result struct {
		ItemID        int32
		TotalQuantity int32
	}

	var rows []result

	err := r.db.WithContext(ctx).
		Model(&movementEntity.Movement{}).
		Select("item_id, SUM(quantity) AS total_quantity").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("item_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	agg := make(map[int32]int32, len(rows))
	for _, row := range rows {
		agg[row.ItemID] = row.TotalQuantity
	}

	return agg, nil
}

func (r *movementRepository) GetSummaryByType(ctx context.Context, from, to time.Time) (map[movementEntity.MovementType]int32, error) {
	type result struct {
		Type          movementEntity.MovementType
		TotalQuantity int32
	}

	var rows []result

	err := r.db.WithContext(ctx).
		Model(&movementEntity.Movement{}).
		Select("movement_type AS type, SUM(quantity) AS total_quantity").
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("movement_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	summary := make(map[movementEntity.MovementType]int32)
	for _, row := range rows {
		summary[row.Type] = row.TotalQuantity
	}

	return summary, nil
}
