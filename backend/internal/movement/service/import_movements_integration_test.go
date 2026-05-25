//go:build integration

package service

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	itemEntity "inventory-movement-processing/internal/item/entity"
	itemRepo "inventory-movement-processing/internal/item/repository/postgres"

	itemSvcPkg "inventory-movement-processing/internal/item/service"

	"inventory-movement-processing/internal/movement/entity"
	movementRepo "inventory-movement-processing/internal/movement/repository/postgres"
	"inventory-movement-processing/pkg/components/gormc"
	"inventory-movement-processing/pkg/components/gormc/dialets"
	"inventory-movement-processing/pkg/components/workerc"

	zaplogger "inventory-movement-processing/pkg/logger/zap"

	"gorm.io/gorm"
)

// =========================================================================
// SETUP
// =========================================================================

var rootDB *gorm.DB

func TestMain(m *testing.M) {
	dbName := fmt.Sprintf("imp_test_service_%d", time.Now().UnixNano())

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

	err = db.AutoMigrate(&itemEntity.Item{}, &entity.Movement{})
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
func setupServiceTest(t *testing.T) (*gorm.DB, MovementService) {
	t.Helper()

	rootDB.Exec("DELETE FROM inventory_movements")
	rootDB.Exec("DELETE FROM inventory_items")

	rootDB.Exec(`
		INSERT INTO inventory_items (id, created_at, updated_at, name, sku, current_stock, low_stock_threshold)
		VALUES (1, NOW(), NOW(), 'Service Integration Item', 'SVC-SKU-01', 100, 10)
	`)
	rootDB.Exec("SELECT setval(pg_get_serial_sequence('inventory_items', 'id'), (SELECT MAX(id) FROM inventory_items))")

	serviceLogger := zaplogger.NewZapLogger()
	log := serviceLogger.GetLogger("integration-test")

	txManager := gormc.NewGormTxManager(rootDB)

	workerPool := workerc.NewPool("test-worker-pool", 5, 50)
	workerPool.InitFlags()
	workerPool.Activate(nil)

	iRepo := itemRepo.NewItemRepository(rootDB)
	mRepo := movementRepo.NewMovementRepository(rootDB)

	iSvc := itemSvcPkg.NewItemService(iRepo, log)
	mSvc := NewMovementService(mRepo, iSvc, txManager, workerPool, log)

	t.Cleanup(func() {
		workerPool.Stop()
	})

	return rootDB, mSvc
}

// =========================================================================
// HELPER: MOCK CSV FILE
// =========================================================================

func createMockCSVFile(t *testing.T, csvContent string) *multipart.FileHeader {
	t.Helper()
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test_import.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte(csvContent))
	writer.Close()

	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(10 << 20)
	if err != nil {
		t.Fatalf("failed to parse form: %v", err)
	}

	return req.MultipartForm.File["file"][0]
}

// =========================================================================
// TEST SUITE: END-TO-END IMPORT BATCH
// =========================================================================

func TestMovementService_ImportBatch_EndToEnd_ShouldProcessSuccessfully(t *testing.T) {
	db, svc := setupServiceTest(t)
	ctx := context.Background()

	csvContent := `external_id,item_id,movement_type,quantity,movement_time,note
					EXT-SVC-01,1,IN,50,2026-05-25T10:00:00Z,Test import 1
					EXT-SVC-02,1,OUT,20,2026-05-25T11:00:00Z,Test import 2
`
	fileHeader := createMockCSVFile(t, csvContent)

	result, err := svc.ImportBatch(ctx, fileHeader)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected Total 2, got %d", result.Total)
	}
	if result.Success != 2 {
		t.Errorf("expected Success 2, got %d", result.Success)
	}

	var count int64
	db.Model(&entity.Movement{}).Where("item_id = ?", 1).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 movements to be saved in DB, got %d", count)
	}

	var item itemEntity.Item
	db.First(&item, 1)
	if item.CurrentStock != 130 {
		t.Errorf("expected final stock to be 130, got %d", item.CurrentStock)
	}
}
