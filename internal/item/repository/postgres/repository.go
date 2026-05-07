package postgres

import "gorm.io/gorm"

type repository struct {
	db *gorm.DB
}

func NewPostgreSQLRepository(db *gorm.DB) *repository {
	return &repository{
		db: db,
	}
}
