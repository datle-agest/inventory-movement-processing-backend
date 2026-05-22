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

// NOTE: What if cache timeout and 100 request hit at once
// NOTE: What if that date is holiday and there is no movement
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

	dateStr := date.Format("2006-01-02")
	cacheKey := fmt.Sprintf("report:top_active:%s", dateStr)

	// 1. Try cache
	cached, found, err := s.cacheStore.Get(ctx, cacheKey)
	if err != nil {
		s.logger.Warnf("Failed to get cache for key %s: %v", cacheKey, err)
	} else if found {
		var cachedReport entity.CachedReport
		if err = json.Unmarshal([]byte(cached), &cachedReport); err == nil {
			isFresh := !isToday || time.Since(cachedReport.GeneratedAt) <= cacheTTLToday
			if isFresh {
				s.logger.Infof("Cache hit (fresh) for daily report key: %s", cacheKey)
				return s.buildResult(ctx, cachedReport.Items, limit, isToday)
			}
			s.logger.Infof("Cache hit but stale for daily report key: %s. Proceeding to regenerate.", cacheKey)
		} else {
			s.logger.Errorf("Failed to unmarshal cache data for key %s: %v", cacheKey, err)
		}
	}

	// 2. If today -> regenerate
	if isToday {
		s.logger.Infof("Report requested for today (%s), triggering summary generation", dateStr)
		if err := s.GenerateDailySummary(ctx, date); err != nil {
			s.logger.Errorf("GenerateDailySummary failed for today: %v", err)
			return nil, common.ErrInternal("failed to process daily summary")
		}
	}

	// 3. Query full ranking from DB
	topItems, err := s.reportRepository.
		ListTopActiveItemsByDate(ctx, date, s.config.GetReportCacheLimit())
	if err != nil {
		s.logger.Errorf("ListTopActiveItemsByDate failed for %s: %v", dateStr, err)
		return nil, common.ErrInternal("cannot query top active items")
	}

	// 4. Check DB staleness
	if isDataStale(topItems, date) {
		s.logger.Infof("DB data is stale for %s, triggering fallback generation", dateStr)
		if err := s.GenerateDailySummary(ctx, date); err != nil {
			s.logger.Errorf("GenerateDailySummary failed during stale fallback: %v", err)
			return nil, common.ErrInternal("failed to process daily summary")
		}

		topItems, err = s.reportRepository.
			ListTopActiveItemsByDate(ctx, date, s.config.GetReportCacheLimit())
		if err != nil {
			s.logger.Errorf("ListTopActiveItemsByDate failed after fallback generation: %v", err)
			return nil, common.ErrInternal("cannot query top active items")
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
		if errCache := s.cacheStore.Set(ctx, cacheKey, string(raw), ttl); errCache != nil {
			s.logger.Warnf("Failed to set cache for key %s: %v", cacheKey, errCache)
		}
	} else {
		s.logger.Errorf("Failed to marshal report for caching: %v", err)
	}

	return s.buildResult(ctx, topItems, limit, isToday)
}

func isDataStale(items []*entity.DailyItemSummary, date time.Time) bool {
	if len(items) == 0 {
		return true
	}

	first := items[0]
	if first.UpdatedAt == nil {
		return true
	}

	now := time.Now()
	isToday := now.Year() == date.Year() && now.Month() == date.Month() && now.Day() == date.Day()

	if isToday {
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		return first.UpdatedAt.Before(startOfDay)
	} else {
		startOfNextDay := time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, date.Location())
		return first.UpdatedAt.Before(startOfNextDay)
	}
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
			s.logger.Errorf("Failed to construct buildResult due to low stock fetch error: %v", err)
			return nil, common.ErrInternal("failed to fetch low stock items")
		}
	}

	return result, nil
}

func (s *reportService) fetchAllLowStockItems(ctx context.Context) ([]*itemEntity.Item, error) {
	items, err := s.itemRepository.ListLowStockItems(ctx)
	if err != nil {
		s.logger.Errorf("ItemRepository.ListLowStockItems failed: %v", err)
		return nil, err
	}

	return items, nil
}
