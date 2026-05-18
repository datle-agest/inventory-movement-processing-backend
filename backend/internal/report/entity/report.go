package entity

import "time"

type CachedReport struct {
	Items       []*DailyItemSummary `json:"items"`
	GeneratedAt time.Time           `json:"generated_at"`
}
