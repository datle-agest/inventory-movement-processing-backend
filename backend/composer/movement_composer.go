package composer

import (
	"inventory-movement-processing/internal/movement/repository/postgres"
	movementService "inventory-movement-processing/internal/movement/service"
	movementHttp "inventory-movement-processing/internal/movement/transport/http"
	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type movementHandler interface {
	GetMovementById() gin.HandlerFunc
}

func ComposeMovementService(serviceCtx sctx.ServiceContext) movementHandler {

	//db := serviceCtx.MustGet(common.KeyComponentPostgres).(*gormc.GormDB).GetDB()

	repo := postgres.NewMovementRepository(new(string))
	uc := movementService.NewMovementService(repo)
	handler := movementHttp.NewHandler(uc)

	return handler
}
