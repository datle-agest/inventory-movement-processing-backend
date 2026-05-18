package configc

import (
	"flag"
	sctx "inventory-movement-processing/pkg/service_context"
)

type config struct {
	id                string
	reportCacheLimit  int
	managerAPIKey     string
	storekeeperAPIKey string
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
	flag.StringVar(
		&c.managerAPIKey,
		"manager-api-key",
		"",
		"API key for manager role",
	)
	flag.StringVar(
		&c.storekeeperAPIKey,
		"storekeeper-api-key",
		"",
		"API key for storekeeper role",
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

func (c *config) GetManagerAPIKey() string {
    return c.managerAPIKey
}

func (c *config) GetStorekeeperAPIKey() string {
    return c.storekeeperAPIKey
}