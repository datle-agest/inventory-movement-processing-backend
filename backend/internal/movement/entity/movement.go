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
	ExternalID string           `json:"external_id"    gorm:"column:external_id;type:varchar(255);uniqueIndex;not null"`
	ItemID     int32            `json:"item_id"        gorm:"column:item_id;not null;index"`
	Item       *itemEntity.Item `json:"item,omitempty" gorm:"foreignKey:ItemID;references:ID"`
	Type       MovementType     `json:"movement_type"  gorm:"column:movement_type;type:varchar(10);not null;index"`
	Quantity   int32            `json:"quantity"       gorm:"column:quantity;not null;check:chk_quantity_nonzero,quantity <> 0"`
	Note       *string          `json:"note"           gorm:"column:note;type:text"`
}

func (Movement) TableName() string {
	return "inventory_movements"
}

func (m *Movement) Validate() error {
	if m.ItemID <= 0 {
		return ErrInvalidItemID
	}
	if m.ExternalID == "" {
		return ErrExternalIDEmpty
	}
	switch m.Type {
	case MovementTypeIn, MovementTypeOut:
		if m.Quantity <= 0 {
			return ErrInvalidQuantity
		}
	case MovementTypeAdjust:
		if m.Quantity == 0 {
			return ErrInvalidQuantity
		}
	default:
		return ErrInvalidType
	}
	return nil
}

type MovementSummary struct {
	TotalInCount     int32
	TotalQtyReceived int32
	TotalOutCount    int32
	TotalQtyIssued   int32
	TotalAdjustCount int32
	TotalAdjustQty   int32
}
