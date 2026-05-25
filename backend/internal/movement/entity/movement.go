package entity

import (
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/core"
	"time"
)

type MovementType string

const (
	MovementTypeIn     MovementType = "IN"
	MovementTypeOut    MovementType = "OUT"
	MovementTypeAdjust MovementType = "ADJUST"
)

type Movement struct {
	core.SQLModel
	ExternalID   string           `json:"external_id"    gorm:"column:external_id;type:varchar(255);uniqueIndex;not null" example:"TXN-120001"`
	ItemID       int32            `json:"item_id"        gorm:"column:item_id;not null;index" example:"1"`
	Item         *itemEntity.Item `json:"item,omitempty" gorm:"foreignKey:ItemID;references:ID" swaggerignore:"true"`
	Type         MovementType     `json:"movement_type"  gorm:"column:movement_type;type:varchar(10);not null;index" example:"IN" enums:"IN,OUT,ADJUST"`
	Quantity     int32            `json:"quantity"       gorm:"column:quantity;not null;check:chk_quantity_nonzero,quantity <> 0" example:"50"`
	MovementTime time.Time        `json:"movement_time"  gorm:"column:movement_time;type:timestamp;not null;index:idx_item_movement_time,type:btree" example:"2026-05-22T08:00:00Z"`
	Note         *string          `json:"note"           gorm:"column:note;type:text" example:"Initial stock import batch 1"`
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
	if m.MovementTime.IsZero() {
		return ErrInvalidMovementTime
	}
	return nil
}

func (m MovementType) IsValid() bool {
	switch m {
	case MovementTypeIn, MovementTypeOut, MovementTypeAdjust:
		return true
	}
	return false
}
