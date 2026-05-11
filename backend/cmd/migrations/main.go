package main

import (
	"log"

	itemEntity "inventory-movement-processing/internal/item/entity"
	movementEntity "inventory-movement-processing/internal/movement/entity"
	reportEntity "inventory-movement-processing/internal/report/entity"

	"inventory-movement-processing/pkg/components/gormc"
	sctx "inventory-movement-processing/pkg/service_context"

	"gorm.io/gorm"
)

type DBProvider interface {
	GetDB() *gorm.DB
}

func RunMigration(db *gorm.DB) error {
	log.Println("Đang tiến hành kiểm tra và chạy Migration...")

	err := db.AutoMigrate(
		&itemEntity.Category{},
		&itemEntity.Item{},
		&movementEntity.Movement{},
		&reportEntity.Report{},
	)

	if err != nil {
		log.Printf("An error occurred during migration: %v\n", err)
		return err
	}

	log.Println("Migration completed successfully!")
	return nil
}

func main() {
	ctx := sctx.NewServiceContext(
		sctx.WithName("migration-tool"),
		sctx.WithComponent(gormc.NewGormDB("gorm", "")),
	)

	if err := ctx.Load(); err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}
	defer ctx.Stop()

	gormComp := ctx.MustGet("gorm").(DBProvider)
	db := gormComp.GetDB()

	if err := RunMigration(db); err != nil {
		log.Fatalf("Program terminated due to migration failure.")
	}
}