package composer

import sctx "inventory-movement-processing/pkg/service_context"

type movementHandler struct {
}

func ComposeMovementService(serviceCtx sctx.ServiceContext) *movementHandler {
	return &movementHandler{}
}
