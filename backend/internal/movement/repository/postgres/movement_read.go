package postgres

import (
	"context"
	"errors"
	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
)

// ErrRecordNotFound là raw error từ DB nào dùng gorm thật: gorm.ErrRecordNotFound sẽ thay thế chỗ này
// Service dùng errors.Is() để check, không compare string
var ErrRecordNotFound = errors.New("record not found")

func (repo *movementRepository) GetMovementById(ctx context.Context, id int) (*entity.Movement, error) {
	// TODO: thay bằng repo.db.WithContext(ctx).First(&m, id).Error khi có gorm

	// giả lập not found và lỗi knoi DB — DB thật sẽ trả ra lỗi tương tự, repo chỉ forward lên
	if id == 404 {
		return nil, ErrRecordNotFound
	}
	if id == 500 {
		return nil, errors.New("database connection failed")
	}

	// giả lập success
	return &entity.Movement{
		ExternalID: "SCAN-00091991",
		ItemID:     012,
		Item:       &itemEntity.Item{Name: "Coca-cola"},
		Type:       entity.MovementTypeIn,
		Quantity:   100,
	}, nil
}
