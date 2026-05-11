package main

import (
	"inventory-movement-processing/pkg/migrations"
	"inventory-movement-processing/pkg/components/gormc"
	sctx "inventory-movement-processing/pkg/service_context"
	"gorm.io/gorm"
	"log"
)

type DBProvider interface {
	GetDB() *gorm.DB
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

	if err := migrations.RunMigration(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}