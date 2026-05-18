package entity

import "time"

// CachedReport wraps the daily item summaries with metadata to support resilience patterns.
// We use this entity instead of storing raw arrays in Redis to decouple
// Physical TTL (Redis eviction) from Logical TTL (business staleness).
type CachedReport struct {
	Items []*DailyItemSummary `json:"items"`

	// GeneratedAt acts as a Logical TTL marker. It enables:
	// 1. Graceful Degradation: If the DB is down, we can fallback to serving slightly stale data instead of returning a 500 error.
	// 2. Cache Stampede Prevention: Allows implementing the Stale-While-Revalidate pattern when multiple requests hit an expired cache.
	// 3. Observability: Provides exact snapshot timing for debugging purposes.
	GeneratedAt time.Time `json:"generated_at"`
}
