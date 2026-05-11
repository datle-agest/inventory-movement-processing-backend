package composer

import sctx "inventory-movement-processing/pkg/service_context"

type reportHandler struct {
}

func ComposeReportService(serviceCtx sctx.ServiceContext) *reportHandler {
	return &reportHandler{}
}
