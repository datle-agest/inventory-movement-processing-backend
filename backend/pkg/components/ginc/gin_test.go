package ginc

import (
	sctx "inventory-movement-processing/pkg/service_context"
	"os"
	"testing"
)

var testServiceCtx sctx.ServiceContext

func TestMain(m *testing.M) {
	_ = os.Setenv("GIN_PORT", "4000")
	_ = os.Setenv("GIN_MODE", "debug")

	testServiceCtx = sctx.NewServiceContext(
		sctx.WithName("test"),
		sctx.WithComponent(NewGin("gin")),
	)

	if err := testServiceCtx.Load(); err != nil {
		panic(err)
	}
	defer testServiceCtx.Stop()

	code := m.Run()
	os.Exit(code)
}

func TestGin_Activate(t *testing.T) {
	g := testServiceCtx.MustGet("gin").(*ginEngine)

	if g.GetRouter() == nil {
		t.Fatal("gin router should not be nil")
	}
}
