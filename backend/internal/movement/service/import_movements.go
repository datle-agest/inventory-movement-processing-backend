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

func (s *service) ProcessOne(ctx context.Context, m *entity.Movement) ProcessStatus {

	s.mu.Lock()
	if s.seen[m.ExternalID] {
		s.mu.Unlock()
		return StatusDuplicate
	}
	s.seen[m.ExternalID] = true
	s.mu.Unlock()

	// validate
	if err := m.Validate(); err != nil {
		return StatusRejected
	}

	time.Sleep(5 * time.Millisecond) // giả lập tg tương tác DB

	return StatusAccepted
}
