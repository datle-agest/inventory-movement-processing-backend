package postgres

import "gorm.io/gorm"

type movementRepository struct {
	db *gorm.DB
}

func NewMovementRepository(db *gorm.DB) *movementRepository {
	return &movementRepository{
		db: db,
	}
}
