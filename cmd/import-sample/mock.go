package main

import "errors"

type movementType string

const (
	movementTypeIn     movementType = "IN"
	movementTypeOut    movementType = "OUT"
	movementTypeAdjust movementType = "ADJUST"
)

type mockMovement struct {
	ID       string
	Name     string
	ItemID   int32
	Type     movementType
	Quantity int32
}

func (m mockMovement) validate() error {
	if m.ID == "" {
		return errors.New("id is required")
	}
	if m.ItemID <= 0 {
		return errors.New("item_id must be > 0")
	}
	if m.Type != movementTypeIn && m.Type != movementTypeOut && m.Type != movementTypeAdjust {
		return errors.New("invalid type")
	}
	if m.Quantity <= 0 {
		return errors.New("quantity must be > 0")
	}
	return nil
}

type result struct {
	AcceptedCount  int32 `json:"accepted_count"`
	RejectedCount  int32 `json:"rejected_count"`
	DuplicateCount int32 `json:"duplicate_count"`
}
