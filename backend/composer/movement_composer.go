package composer

import (
	"inventory-movement-processing/common"

	itemPostgres "inventory-movement-processing/internal/item/repository/postgres"
	movementPostgres "inventory-movement-processing/internal/movement/repository/postgres"

	movementService "inventory-movement-processing/internal/movement/service"
	movementHttp "inventory-movement-processing/internal/movement/transport/http"

	sctx "inventory-movement-processing/pkg/service_context"

	"github.com/gin-gonic/gin"
)

type movementHandler interface {
	GetMovementsByItemID() gin.HandlerFunc
}

func ComposeMovementService(serviceCtx sctx.ServiceContext) movementHandler {
	db := serviceCtx.MustGet(common.KeyComponentPostgres).(common.DBProvider).GetDB()

	movementRepo := movementPostgres.NewMovementRepository(db)
	itemRepo := itemPostgres.NewItemRepository(db)

	uc := movementService.NewMovementService(
		movementRepo,
		itemRepo,
	)

	handler := movementHttp.NewHandler(uc)

	return handler
}
