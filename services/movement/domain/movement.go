package domain

import "inventory-movement-processing/pkg/core"

type MovementType string

const (
	MovementTypeIn     MovementType = "IN"
	MovementTypeOut    MovementType = "OUT"
	MovementTypeAdjust MovementType = "ADJUST"
)

type Movement struct {
	core.SQLModel
	Name     string       `json:"name"`
	ItemID   int32        `json:"item_id"`
	Type     MovementType `json:"movement_type"`
	Quantity int32        `json:"quantity"`
}

func (Movement) TableName() string {
	return "inventory_movements"
}
