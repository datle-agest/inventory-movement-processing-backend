package configc

import (
	"flag"
	sctx "inventory-movement-processing/pkg/service_context"
)

type config struct {
	id               string
	reportCacheLimit int
}

func NewConfigComponent(id string) *config {
	return &config{id: id}
}

func (c *config) InitFlags() {
	flag.IntVar(
		&c.reportCacheLimit,
		"top k report cache limit",
		100,
		"cache top k (default: 100)",
	)
}

func (c *config) ID() string {
	return c.id
}

func (c *config) Activate(_ sctx.ServiceContext) error {
	return nil
}

func (c *config) Stop() error {
	return nil
}

func (c *config) GetReportCacheLimit() int {
	return c.reportCacheLimit
}
