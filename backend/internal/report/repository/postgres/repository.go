package postgres

import "gorm.io/gorm"

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *reportRepository {
	return &reportRepository{
		db: db,
	}
}
