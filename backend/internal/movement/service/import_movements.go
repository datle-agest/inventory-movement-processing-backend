package service

import (
	"context"
	"inventory-movement-processing/internal/movement/entity"
	"time"
)

type ProcessStatus string

const (
	StatusAccepted  ProcessStatus = "accepted"
	StatusRejected  ProcessStatus = "rejected"
	StatusDuplicate ProcessStatus = "duplicate"
)

func (uc *service) ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus {

	uc.mu.Lock()
	if uc.seen[m.ExternalID] {
		uc.mu.Unlock()
		return StatusDuplicate
	}
	uc.seen[m.ExternalID] = true
	uc.mu.Unlock()

	// validate
	if err := m.Validate(); err != nil {
		return StatusRejected
	}

	time.Sleep(5 * time.Millisecond) // giả lập tg tương tác DB

	return StatusAccepted
}
