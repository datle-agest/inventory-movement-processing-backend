package service

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory-movement-processing/common"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/report/entity"
	"time"
)

const cacheTTLToday = 1 * time.Minute

func (s *reportService) GetDailyReport(
	ctx context.Context,
	date time.Time,
	limit int,
) (*entity.TopActiveItemsResult, error) {

	if limit <= 0 {
		limit = 5
	}

	now := time.Now()
	isToday :=
		now.Year() == date.Year() &&
			now.Month() == date.Month() &&
			now.Day() == date.Day()

	cacheKey := fmt.Sprintf(
		"report:top_active:%s",
		date.Format("2006-01-02"),
	)

	// 1. Try cache
	cached, found, err := s.cacheStore.Get(ctx, cacheKey)
	if err == nil && found {
		var cachedReport entity.CachedReport

		if err = json.Unmarshal([]byte(cached), &cachedReport); err == nil {
			isFresh := !isToday || time.Since(cachedReport.GeneratedAt) <= cacheTTLToday
			if isFresh {
				return s.buildResult(ctx, cachedReport.Items, limit, isToday)
			}
		}
	}

	// 2. If today -> regenerate
	if isToday {
		if err := s.GenerateDailySummary(ctx, date); err != nil {
			return nil, common.ErrInternal(err.Error())
		}
	}

	// 3. Query full ranking from DB
	topItems, err := s.reportRepository.
		ListTopActiveItemsByDate(ctx, date, s.config.GetReportCacheLimit())
	if err != nil {
		return nil, common.ErrInternal("cannot query top active items")
	}

	// 4. Check DB staleness
	if isDataStale(topItems, date) {
		if err := s.GenerateDailySummary(ctx, date); err != nil {
			return nil, common.ErrInternal(err.Error())
		}

		topItems, err = s.reportRepository.
			ListTopActiveItemsByDate(ctx, date, s.config.GetReportCacheLimit())
		if err != nil {
			return nil, err
		}
	}

	// 5. Cache full list with timestamp
	cachedReport := entity.CachedReport{
		Items:       topItems,
		GeneratedAt: now,
	}
	if raw, err := json.Marshal(cachedReport); err == nil {
		ttl := 24 * time.Hour
		if isToday {
			ttl = cacheTTLToday
		}
		_ = s.cacheStore.Set(ctx, cacheKey, string(raw), ttl)
	}

	return s.buildResult(ctx, topItems, limit, isToday)
}

func isDataStale(items []*entity.DailyItemSummary, date time.Time) bool {
	if len(items) == 0 {
		return true
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	first := items[0]
	if first.UpdatedAt == nil {
		return true
	}

	return first.UpdatedAt.Before(startOfDay)
}

func (s *reportService) buildResult(
	ctx context.Context,
	topItems []*entity.DailyItemSummary,
	limit int,
	isToday bool,
) (*entity.TopActiveItemsResult, error) {

	if limit > len(topItems) {
		limit = len(topItems)
	}

	result := &entity.TopActiveItemsResult{
		TopItems: topItems[:limit],
	}

	if isToday {
		var err error
		result.LowStockItems, err = s.fetchAllLowStockItems(ctx)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (s *reportService) fetchAllLowStockItems(ctx context.Context) ([]*itemEntity.Item, error) {
	items, err := s.itemRepository.ListLowStockItems(ctx)
	if err != nil {
		return nil, common.ErrInternal(err.Error())
	}

	return items, nil
}
