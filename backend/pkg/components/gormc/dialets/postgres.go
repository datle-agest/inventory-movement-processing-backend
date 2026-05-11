package dialets

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func PostgresDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
