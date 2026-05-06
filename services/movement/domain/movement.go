package domain

import (
	"inventory-movement-processing/pkg/core"
	itemDomain "inventory-movement-processing/services/item/domain"
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
