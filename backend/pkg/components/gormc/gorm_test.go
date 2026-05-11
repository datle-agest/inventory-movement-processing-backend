package gormc

import (
	sctx "inventory-movement-processing/pkg/service_context"
	"os"
	"testing"
)

var testServiceCtx sctx.ServiceContext

func TestMain(m *testing.M) {
	_ = os.Setenv("DB_DRIVER", "postgres")
    _ = os.Setenv("DB_DSN", "host=localhost user=postgres password=123456 dbname=my_db port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh")

	testServiceCtx = sctx.NewServiceContext(
		sctx.WithName("test"),
		sctx.WithComponent(NewGormDB("gorm", "")),
	)

	if err := testServiceCtx.Load(); err != nil {
		panic(err)
	}

	defer testServiceCtx.Stop()

	code := m.Run()
	os.Exit(code)
}

func TestGormDB_Connect_Success(t *testing.T) {
	gormComp := testServiceCtx.MustGet("gorm").(*gormDB)

	db := gormComp.db
	if db == nil {
		t.Fatal("gorm db should not be nil")
	}
}

type TestUser struct {
	ID   uint
	Name string
}

func TestGormDB_InsertFind_Success(t *testing.T) {
	gormComp := testServiceCtx.MustGet("gorm").(*gormDB)
	db := gormComp.GetDB()

	err := db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatal(err)
	}

	user := TestUser{Name: "dat"}

	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	var found TestUser
	if err := db.First(&found, "name = ?", "dat").Error; err != nil {
		t.Fatal(err)
	}
}
