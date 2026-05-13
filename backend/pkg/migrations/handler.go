package migrations

import (
	itemEntity "inventory-movement-processing/internal/item/entity"
	importBatchEntity "inventory-movement-processing/internal/movement/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	dailyItemSummaryEntity "inventory-movement-processing/internal/report/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"

	"log"

	"gorm.io/gorm"
)

func RunMigration(db *gorm.DB) error {
	log.Println("Checking and running migrations...")

	err := db.AutoMigrate(
		&itemEntity.Item{},
		&movementEntity.Movement{},
		&reportEntity.Report{},
		&importBatchEntity.ImportBatch{},
		&dailyItemSummaryEntity.DailyItemSummary{},
	)

	if err != nil {
		log.Printf("Migration error: %v\n", err)
		return err
	}

	log.Println("Migration completed successfully!")
	return nil
}
