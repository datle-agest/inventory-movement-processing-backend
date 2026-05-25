

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/report/entity"
	"inventory-movement-processing/pkg/components/gormc/dialets"

	"gorm.io/gorm"
)

// =========================================================================
// ENVIRONMENT SETUP
// =========================================================================

var rootDB *gorm.DB

func TestMain(m *testing.M) {
	dbName := fmt.Sprintf("imp_test_report_%d", time.Now().UnixNano())

	baseDSN := "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable"
	baseDB, err := dialets.PostgresDB(baseDSN)
	if err != nil {
		fmt.Printf("Cannot connect to base Postgres: %v\n", err)
		os.Exit(1)
	}

	baseDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))

	testDSN := fmt.Sprintf("host=localhost user=postgres password=123456 dbname=%s port=5432 sslmode=disable", dbName)
	db, err := dialets.PostgresDB(testDSN)
	if err != nil {
		fmt.Printf("Cannot connect to test DB %s: %v\n", dbName, err)
		os.Exit(1)
	}

	err = db.AutoMigrate(&itemEntity.Item{}, &entity.DailyItemSummary{})
	if err != nil {
		fmt.Printf("Failed to migrate test DB: %v\n", err)
		os.Exit(1)
	}

	rootDB = db
	code := m.Run()

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	baseDB.Exec(fmt.Sprintf("DROP DATABASE %s;", dbName))

	os.Exit(code)
}

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	if rootDB == nil {
		t.Skip("database not available")
	}

	rootDB.Exec("DELETE FROM daily_item_summary") 
	rootDB.Exec("DELETE FROM inventory_items")

	return rootDB
}

// =========================================================================
// TEST SUITE
// =========================================================================

func TestReportRepo_UpsertDailyItemSummary_ShouldInsertAndUpdate(t *testing.T) {
	db := setupDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	items := []itemEntity.Item{
		{Name: "Item 1", SKU: "SKU-01"},
		{Name: "Item 2", SKU: "SKU-02"},
	}
	items[0].ID = 1
	items[1].ID = 2
	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("failed to seed items: %v", err)
	}

	today := time.Now().Truncate(24 * time.Hour)

	initialData := []*entity.DailyItemSummary{
		{ItemID: 1, SummaryDate: today, TotalIn: 10, TotalOut: 5, TotalAdjust: 0},
	}
	err := repo.UpsertDailyItemSummary(ctx, initialData)
	if err != nil {
		t.Fatalf("expected no error on initial insert, got %v", err)
	}

	var count int64
	db.Model(&entity.DailyItemSummary{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 record after insert, got %d", count)
	}

	updateData := []*entity.DailyItemSummary{
		{ItemID: 1, SummaryDate: today, TotalIn: 20, TotalOut: 15, TotalAdjust: 2}, // Should update
		{ItemID: 2, SummaryDate: today, TotalIn: 100, TotalOut: 0, TotalAdjust: 0}, // Should insert
	}
	
	err = repo.UpsertDailyItemSummary(ctx, updateData)
	if err != nil {
		t.Fatalf("expected no error on upsert, got %v", err)
	}

	db.Model(&entity.DailyItemSummary{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 records after upsert, got %d", count)
	}

	var updatedItem entity.DailyItemSummary
	db.Where("item_id = ? AND summary_date = ?", 1, today).First(&updatedItem)
	if updatedItem.TotalIn != 20 || updatedItem.TotalOut != 15 {
		t.Errorf("expected updated values (20, 15), got (%d, %d)", updatedItem.TotalIn, updatedItem.TotalOut)
	}
}

func TestReportRepo_ListTopActiveItemsByDate_ShouldReturnSortedAndLimited(t *testing.T) {
	db := setupDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	items := []itemEntity.Item{
		{Name: "Item 1", SKU: "SKU-01"},
		{Name: "Item 2", SKU: "SKU-02"},
		{Name: "Item 3", SKU: "SKU-03"},
	}
	items[0].ID = 1
	items[1].ID = 2
	items[2].ID = 3

	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("failed to seed items: %v", err)
	}

	testDate := time.Date(2026, 5, 25, 0, 0, 0, 0, time.Local)

	summaries := []entity.DailyItemSummary{
		{ItemID: 1, SummaryDate: testDate, TotalIn: 10, TotalOut: 0, TotalAdjust: 0},  // Total = 10
		{ItemID: 2, SummaryDate: testDate, TotalIn: 20, TotalOut: 30, TotalAdjust: 0}, // Total = 50 (Expected Top 1)
		{ItemID: 3, SummaryDate: testDate, TotalIn: 15, TotalOut: 10, TotalAdjust: 5}, // Total = 30 (Expected Top 2)
	}
	if err := db.Create(&summaries).Error; err != nil {
		t.Fatalf("failed to seed summaries: %v", err)
	}

	results, err := repo.ListTopActiveItemsByDate(ctx, testDate, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results due to limit, got %d", len(results))
	}

	if results[0].ItemID != 2 {
		t.Errorf("expected Top 1 to be ItemID 2, got %d", results[0].ItemID)
	}
	
	if results[1].ItemID != 3 {
		t.Errorf("expected Top 2 to be ItemID 3, got %d", results[1].ItemID)
	}

	if results[0].Item.Name != "Item 2" {
		t.Errorf("expected preloaded item name to be 'Item 2', got '%s'", results[0].Item.Name)
	}
}