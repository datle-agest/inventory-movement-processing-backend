package dialets

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// host=localhost user=postgres password=12345 dbname=mydb port=5432 sslmode=disable
func PostgresDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
}
