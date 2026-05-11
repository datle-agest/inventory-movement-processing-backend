package postgres

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
