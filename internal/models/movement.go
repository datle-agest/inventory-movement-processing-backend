package domain

import (
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
	ItemID   int32        `json:"item_id"`
	Item     *Item        `json:"item,omitempty"`
	Type     MovementType `json:"movement_type"`
	Quantity int32        `json:"quantity"`
}

func (Movement) TableName() string {
	return "inventory_movements"
}
