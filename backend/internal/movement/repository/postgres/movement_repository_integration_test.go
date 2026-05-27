//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	"inventory-movement-processing/internal/movement/entity"
	"inventory-movement-processing/pkg/components/gormc"
	"inventory-movement-processing/pkg/components/gormc/dialets"
	"inventory-movement-processing/pkg/core"

	"gorm.io/gorm"
)

// =========================================================================
// SETUP
// =========================================================================

var rootDB *gorm.DB

func TestMain(m *testing.M) {
	dbName := fmt.Sprintf("imp_test_movement_%d", time.Now().UnixNano())

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

	err = db.AutoMigrate(
		&itemEntity.Item{},
		&entity.Movement{},
	)
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

	rootDB.Exec(`
		INSERT INTO inventory_items (id, created_at, updated_at, name, sku, current_stock, low_stock_threshold)
		VALUES
			(1, NOW(), NOW(), 'Test Item 1', 'TEST-SKU-001', 100, 10),
			(2, NOW(), NOW(), 'Test Item 2', 'TEST-SKU-002', 100, 10),
			(3, NOW(), NOW(), 'Test Item 3', 'TEST-SKU-003', 100, 10)
		ON CONFLICT DO NOTHING
	`)

	rootDB.Exec("DELETE FROM inventory_movements")
	t.Cleanup(func() {
		rootDB.Exec("DELETE FROM inventory_movements")
	})
	
	return rootDB
}

func newRepo(db *gorm.DB) *movementRepository {
	return NewMovementRepository(db)
}

func ptr(s string) *string { return &s }

func makeMovement(externalID string, itemID int32, mType entity.MovementType, qty int32, t time.Time) *entity.Movement {
	return &entity.Movement{
		ExternalID:   externalID,
		ItemID:       itemID,
		Type:         mType,
		Quantity:     qty,
		MovementTime: t,
		Note:         ptr("test note"),
	}
}

// =========================================================================
// TEST SUITE: Create
// =========================================================================

// TEST CASE 1: Create valid movement -> should persist to DB
func TestMovementRepo_Create_ValidMovement_ShouldPersist(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	m := makeMovement("EXT-001", 1, entity.MovementTypeIn, 10, time.Now())

	err := repo.Create(ctx, m)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if m.ID == 0 {
		t.Error("expected ID to be set after create")
	}

	var found entity.Movement
	db.First(&found, "external_id = ?", "EXT-001")
	if found.ExternalID != "EXT-001" {
		t.Errorf("expected to find movement EXT-001, got: %s", found.ExternalID)
	}
	if found.Quantity != 10 {
		t.Errorf("expected quantity=10, got %d", found.Quantity)
	}
}

// TEST CASE 2: Create duplicate external_id -> should return ErrDuplicateMovement
func TestMovementRepo_Create_DuplicateExternalID_ShouldReturnErrDuplicate(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	m1 := makeMovement("EXT-DUP", 1, entity.MovementTypeIn, 10, time.Now())
	m2 := makeMovement("EXT-DUP", 2, entity.MovementTypeIn, 5, time.Now())

	if err := repo.Create(ctx, m1); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	err := repo.Create(ctx, m2)
	if err == nil {
		t.Fatal("expected ErrDuplicateMovement, got nil")
	}
	if err != itemEntity.ErrDuplicateMovement {
		t.Errorf("expected ErrDuplicateMovement, got: %v", err)
	}
}

// TEST CASE 3: Create inside transaction -> commit should persist
func TestMovementRepo_Create_InsideTransaction_ShouldPersistOnCommit(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()
	txManager := gormc.NewGormTxManager(db)

	m := makeMovement("EXT-TX-OK", 1, entity.MovementTypeIn, 20, time.Now())

	err := txManager.WithTx(ctx, func(txCtx context.Context) error {
		return repo.Create(txCtx, m)
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var found entity.Movement
	db.First(&found, "external_id = ?", "EXT-TX-OK")
	if found.ExternalID != "EXT-TX-OK" {
		t.Error("expected movement to be persisted after commit")
	}
}

// TEST CASE 4: Create inside transaction -> rollback should not persist
func TestMovementRepo_Create_InsideTransaction_ShouldRollbackOnError(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()
	txManager := gormc.NewGormTxManager(db)

	m := makeMovement("EXT-TX-ROLLBACK", 1, entity.MovementTypeIn, 20, time.Now())

	_ = txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Create(txCtx, m); err != nil {
			return err
		}
		return gorm.ErrInvalidTransaction // Force rollback
	})

	var count int64
	db.Model(&entity.Movement{}).Where("external_id = ?", "EXT-TX-ROLLBACK").Count(&count)
	if count != 0 {
		t.Error("expected movement to NOT be persisted after rollback")
	}
}

// =========================================================================
// TEST SUITE: GetMovementsByItemID
// =========================================================================

// TEST CASE 5: No movements for item -> should return empty slice
func TestMovementRepo_GetMovementsByItemID_NoRows_ShouldReturnEmpty(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	paging := &core.Pagination{Page: 1, Limit: 10}
	results, err := repo.GetMovementsByItemID(ctx, 999, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	if paging.Total != 0 {
		t.Errorf("expected Total=0, got %d", paging.Total)
	}
}

// TEST CASE 6: Multiple movements for item -> should return correct count and set Total
func TestMovementRepo_GetMovementsByItemID_MultipleRows_ShouldReturnAll(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	now := time.Now()
	repo.Create(ctx, makeMovement("EXT-A", 1, entity.MovementTypeIn, 10, now))
	repo.Create(ctx, makeMovement("EXT-B", 1, entity.MovementTypeOut, 5, now.Add(time.Minute)))
	repo.Create(ctx, makeMovement("EXT-C", 2, entity.MovementTypeIn, 3, now.Add(2*time.Minute)))

	paging := &core.Pagination{Page: 1, Limit: 10}
	results, err := repo.GetMovementsByItemID(ctx, 1, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
    }
	if len(results) != 2 {
		t.Errorf("expected 2 results for item_id=1, got %d", len(results))
	}
	if paging.Total != 2 {
		t.Errorf("expected Total=2, got %d", paging.Total)
	}
}

// TEST CASE 7: Pagination -> page 2 should return correct offset slice
func TestMovementRepo_GetMovementsByItemID_Pagination_ShouldOffset(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 5; i++ {
		repo.Create(ctx, makeMovement(
			"EXT-PAGE-"+string(rune('A'+i)),
			1,
			entity.MovementTypeIn,
			int32(i+1),
			now.Add(time.Duration(i)*time.Minute),
		))
	}

	paging := &core.Pagination{Page: 2, Limit: 2}
	results, err := repo.GetMovementsByItemID(ctx, 1, paging)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results on page 2, got %d", len(results))
	}
	if paging.Total != 5 {
		t.Errorf("expected Total=5, got %d", paging.Total)
	}
}

// =========================================================================
// TEST SUITE: AggregateDailyItemSummaryFromMovement
// =========================================================================

// TEST CASE 8: No movements in range -> should return empty
func TestMovementRepo_AggregateDailyItemSummary_NoRows_ShouldReturnEmpty(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	start := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	results, err := repo.AggregateDailyItemSummaryFromMovement(ctx, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TEST CASE 9: Movements in range -> should aggregate IN/OUT/ADJUST correctly per item
func TestMovementRepo_AggregateDailyItemSummary_ShouldSumCorrectly(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	day := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)

	m1 := makeMovement("E1", 1, entity.MovementTypeIn, 10, day)
	repo.Create(ctx, m1)
	db.Model(m1).Update("created_at", day)

	m2 := makeMovement("E2", 1, entity.MovementTypeIn, 20, day.Add(time.Hour))
	repo.Create(ctx, m2)
	db.Model(m2).Update("created_at", day.Add(time.Hour))

	m3 := makeMovement("E3", 1, entity.MovementTypeOut, 10, day.Add(2*time.Hour))
	repo.Create(ctx, m3)
	db.Model(m3).Update("created_at", day.Add(2*time.Hour))

	m4 := makeMovement("E4", 1, entity.MovementTypeAdjust, 5, day.Add(3*time.Hour))
	repo.Create(ctx, m4)
	db.Model(m4).Update("created_at", day.Add(3*time.Hour))

	m5 := makeMovement("E5", 2, entity.MovementTypeIn, 50, day)
	repo.Create(ctx, m5)
	db.Model(m5).Update("created_at", day)

	start := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	results, err := repo.AggregateDailyItemSummaryFromMovement(ctx, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 summary rows, got %d", len(results))
	}

	summaryMap := make(map[int32]*struct{ in, out, adjust int32 })
	for _, r := range results {
		summaryMap[r.ItemID] = &struct{ in, out, adjust int32 }{
			in: r.TotalIn, out: r.TotalOut, adjust: r.TotalAdjust,
		}
	}

	item1 := summaryMap[1]
	if item1 == nil {
		t.Fatal("expected summary for item_id=1")
	}
	if item1.in != 30 {
		t.Errorf("item1: expected TotalIn=30, got %d", item1.in)
	}
	if item1.out != 10 {
		t.Errorf("item1: expected TotalOut=10, got %d", item1.out)
	}
	if item1.adjust != 5 {
		t.Errorf("item1: expected TotalAdjust=5, got %d", item1.adjust)
	}

	item2 := summaryMap[2]
	if item2 == nil {
		t.Fatal("expected summary for item_id=2")
	}
	if item2.in != 50 {
		t.Errorf("item2: expected TotalIn=50, got %d", item2.in)
	}
}

// TEST CASE 10: Movements outside date range -> should not be included
func TestMovementRepo_AggregateDailyItemSummary_OutOfRange_ShouldExclude(t *testing.T) {
	db := setupDB(t)
	repo := newRepo(db)
	ctx := context.Background()

	inRange := time.Date(2026, 5, 25, 8, 0, 0, 0, time.UTC)
	outOfRange := time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC)

	m1 := makeMovement("IN-RANGE", 1, entity.MovementTypeIn, 100, inRange)
	repo.Create(ctx, m1)
	db.Model(m1).Update("created_at", inRange)

	m2 := makeMovement("OUT-RANGE", 1, entity.MovementTypeIn, 999, outOfRange)
	repo.Create(ctx, m2)
	db.Model(m2).Update("created_at", outOfRange)

	start := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)

	results, err := repo.AggregateDailyItemSummaryFromMovement(ctx, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].TotalIn != 100 {
		t.Errorf("expected TotalIn=100 (only in-range), got %d", results[0].TotalIn)
	}
}