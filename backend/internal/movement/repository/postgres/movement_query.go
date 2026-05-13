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
		Select("item_id, SUM(ABS(quantity)) AS total_quantity").
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

func (r *movementRepository) GetSummaryByType(ctx context.Context, from, to time.Time) (*movementEntity.MovementSummary, error) {
	type result struct {
		MovementType     movementEntity.MovementType
		Count            int32
		TotalPositiveQty int32
		TotalNegativeQty int32
	}

	var rows []result

	err := r.db.WithContext(ctx).
		Model(&movementEntity.Movement{}).
		Select(`
			movement_type,
			COUNT(*) AS count,
			COALESCE(SUM(CASE WHEN quantity > 0 THEN quantity ELSE 0 END), 0)      AS total_positive_qty,
			COALESCE(SUM(CASE WHEN quantity < 0 THEN ABS(quantity) ELSE 0 END), 0) AS total_negative_qty
		`).
		Where("created_at >= ? AND created_at < ?", from, to).
		Group("movement_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	summary := &movementEntity.MovementSummary{}

	for _, row := range rows {
		switch row.MovementType {
		case movementEntity.MovementTypeIn:
			summary.TotalInCount = row.Count
			summary.TotalQtyReceived += row.TotalPositiveQty

		case movementEntity.MovementTypeOut:
			summary.TotalOutCount = row.Count
			summary.TotalQtyIssued += row.TotalPositiveQty

		case movementEntity.MovementTypeAdjust:
			summary.TotalAdjustCount = row.Count
			summary.TotalQtyReceived += row.TotalPositiveQty
			summary.TotalQtyIssued += row.TotalNegativeQty
		}
	}

	return summary, nil
}

func (r *movementRepository) GetMovementsByItemID(ctx context.Context, itemId int) ([]movementEntity.Movement, error) {
	var movements []movementEntity.Movement
	err := r.db.WithContext(ctx).Where("item_id = ?", itemId).Find(&movements).Error
	return movements, err
}
