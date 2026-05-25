//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/pkg/components/gormc/dialets"

	"gorm.io/gorm"
)

// =========================================================================
// SETUP
// =========================================================================

var rootDB *gorm.DB

func TestMain(m *testing.M) {
	dbName := fmt.Sprintf("imp_test_item_%d", time.Now().UnixNano())

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

	err = db.AutoMigrate(&entity.Item{})
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

	rootDB.Exec("DELETE FROM inventory_items")

	rootDB.Exec(`
		INSERT INTO inventory_items (id, created_at, updated_at, name, sku, current_stock, low_stock_threshold)
		VALUES (1, NOW(), NOW(), 'Test Item', 'SKU-001', 100, 10)
	`)

	rootDB.Exec("SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items))")

	t.Cleanup(func() {
		rootDB.Exec("DELETE FROM inventory_items")
	})

	return rootDB
}

func newRepo(db *gorm.DB) *repository {
	return NewItemRepository(db)
}

// =========================================================================
// TEST SUITE: CRITICAL PATH (Stock Update & Locking)
// =========================================================================

// TEST CASE 1: GetItemForUpdate -> Should lock and return the item correctly
func TestItemRepo_GetItemForUpdate_ShouldReturnItem(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	err := db.Transaction(func(tx *gorm.DB) error {
		txRepo := newRepo(tx)
		item, err := txRepo.GetItemForUpdate(ctx, 1)

		if err != nil {
			return err
		}
		if item == nil {
			t.Fatalf("expected item to be returned, got nil")
		}
		if item.CurrentStock != 100 {
			t.Errorf("expected stock 100, got %d", item.CurrentStock)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error during GetItemForUpdate: %v", err)
	}
}

// TEST CASE 2: UpdateStock -> Should update absolute stock value correctly
func TestItemRepo_UpdateStock_ShouldSetNewStock(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	err := repo.UpdateStock(ctx, 1, 250)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var item entity.Item
	db.First(&item, 1)
	if item.CurrentStock != 250 {
		t.Errorf("expected stock to be 250, got %d", item.CurrentStock)
	}
}

// =========================================================================
// TEST SUITE: OTHER CORE METHODS
// =========================================================================

// TEST CASE 3: CreateItem -> Should persist and handle duplicates
func TestItemRepo_CreateItem_ShouldPersistAndHandleDuplicates(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	newItem := entity.Item{
		Name:              "New Item",
		SKU:               "NEW-SKU",
		CurrentStock:      50,
		LowStockThreshold: 5,
	}

	created, err := repo.CreateItem(ctx, newItem)
	if err != nil {
		t.Fatalf("expected no error on first create, got: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected ID to be assigned")
	}

	duplicateItem := entity.Item{Name: "Duplicate", SKU: "NEW-SKU"}
	_, err = repo.CreateItem(ctx, duplicateItem)
	if err != entity.ErrItemDuplicated {
		t.Errorf("expected ErrItemDuplicated for duplicate SKU, got: %v", err)
	}
}

// TEST CASE 4: ListLowStockItems -> Should return items below threshold
func TestItemRepo_ListLowStockItems_ShouldReturnCorrectly(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	db.Exec(`INSERT INTO inventory_items (id, created_at, updated_at, name, sku, current_stock, low_stock_threshold)
             VALUES (2, NOW(), NOW(), 'Low Item', 'LOW-001', 3, 10)`)

	db.Exec(`INSERT INTO inventory_items (id, created_at, updated_at, name, sku, current_stock, low_stock_threshold)
             VALUES (3, NOW(), NOW(), 'Empty Item', 'EMP-001', 0, 5)`)

	items, err := repo.ListLowStockItems(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 low stock items, got %d", len(items))
	}

	if items[0].ID != 2 || items[1].ID != 3 {
		t.Errorf("items are not sorted correctly by severity")
	}
}
