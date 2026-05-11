package entity

import (
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
)

type MovementType string

const (
	MovementTypeIn     MovementType = "IN"
	MovementTypeOut    MovementType = "OUT"
	MovementTypeAdjust MovementType = "ADJUST"
)

type Movement struct {
	core.SQLModel
	ExternalID string           `json:"external_id"` // id này do scanner scan item mục đích là detect duplicate do chưa có DB đang inmemory
	ItemID     int32            `json:"item_id"`
	Item       *itemEntity.Item `json:"item,omitempty"`
	Type       MovementType     `json:"movement_type"`
	Quantity   int32            `json:"quantity"`
}

func (Movement) TableName() string {
	return "inventory_movements"
}

func (m *Movement) Validate() error {
	if m.ItemID <= 0 {
		return ErrInvalidItemID
	}

	if m.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	switch m.Type {
	case MovementTypeIn, MovementTypeOut, MovementTypeAdjust:
		return nil
	default:
		return ErrInvalidType
	}
}
