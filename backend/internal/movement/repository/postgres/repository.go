package postgres

import (
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewMovementRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
