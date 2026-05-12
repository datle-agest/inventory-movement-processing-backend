package service

import (
	"context"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/report/entity"
	"sort"
	"time"
)

func (s *reportService) CreateReport(ctx context.Context, date time.Time) (*entity.Report, error) {
	from, to := normalizeDate(date)

	summary, err := s.movementRepository.GetSummaryByType(ctx, from, to)
	if err != nil {
		return nil, err
	}

	agg, err := s.movementRepository.GetAggregatedByItem(ctx, from, to)
	if err != nil {
		return nil, err
	}

	top5 := calculateTop5(agg)

	items, err := s.itemRepository.GetLowStockItems(ctx)
	if err != nil {
		return nil, err
	}
	lowStock := mapLowStockItems(items)

	report := &entity.Report{
		ReportDate:            date,
		TotalInCount:          summary.TotalInCount,
		TotalOutCount:         summary.TotalOutCount,
		TotalAdjustCount:      summary.TotalAdjustCount,
		TotalQuantityReceived: summary.TotalQtyReceived, // IN qty + ADJUST dương
		TotalQuantityIssued:   summary.TotalQtyIssued,   // OUT qty + ADJUST âm
		Top5ActiveItem:        top5,
		LowStockItem:          lowStock,
	}

	return report, s.reportRepository.CreateReport(ctx, report)
}

func normalizeDate(t time.Time) (start time.Time, end time.Time) {
	utc := t.UTC()

	start = time.Date(
		utc.Year(), utc.Month(), utc.Day(),
		0, 0, 0, 0,
		time.UTC,
	)
	end = start.Add(24 * time.Hour)

	return start, end
}

func calculateTop5(data map[int32]int32) []entity.ReportItem {
	result := make([]entity.ReportItem, 0, len(data))

	for id, qty := range data {
		result = append(result, entity.ReportItem{
			ItemID:   id,
			Quantity: qty,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Quantity == result[j].Quantity {
			return result[i].ItemID < result[j].ItemID
		}
		return result[i].Quantity > result[j].Quantity
	})

	if len(result) > 5 {
		result = result[:5]
	}

	return result
}

func mapLowStockItems(items []itemEntity.Item) []entity.ReportItem {
	result := make([]entity.ReportItem, 0, len(items))

	for _, it := range items {
		result = append(result, entity.ReportItem{
			ItemID:   it.ID,
			Name:     it.Name,
			Quantity: it.CurrentStock,
		})
	}

	return result
}
