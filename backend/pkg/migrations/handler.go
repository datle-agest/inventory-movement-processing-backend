package migrations // Importable package name

import (
	"log"
	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"
	"gorm.io/gorm"
)

// RunMigration is the core logic shared between your app startup and the CLI tool
func RunMigration(db *gorm.DB) error {
	log.Println("Checking and running migrations...")

	err := db.AutoMigrate(
		&itemEntity.Category{},
		&itemEntity.Item{},
		&movementEntity.Movement{},
		&reportEntity.Report{},
	)

	if err != nil {
		log.Printf("Migration error: %v\n", err)
		return err
	}

	log.Println("Migration completed successfully!")
	return nil
}