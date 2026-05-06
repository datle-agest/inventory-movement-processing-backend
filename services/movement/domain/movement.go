package domain

import (
	"errors"
	"inventory-movement-processing/pkg/core"
	itemDomain "inventory-movement-processing/services/item/domain"
)

var (
	ErrInvalidItemID   = errors.New("item_id must be greater than 0")
	ErrInvalidQuantity = errors.New("quantity must be greater than 0")
	ErrInvalidType     = errors.New("movement_type must be IN, OUT or ADJUST")
)

type MovementType string

const (
	MovementTypeIn     MovementType = "IN"
	MovementTypeOut    MovementType = "OUT"
	MovementTypeAdjust MovementType = "ADJUST"
)

type Movement struct {
	core.SQLModel
	ItemID   int32            `json:"item_id"`
	Item     *itemDomain.Item `json:"item,omitempty"`
	Type     MovementType     `json:"movement_type"`
	Quantity int32            `json:"quantity"`
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