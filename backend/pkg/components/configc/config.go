package configc

import (
	"flag"
	sctx "inventory-movement-processing/pkg/service_context"
)

type ConfigComponent interface {
	GetStorekeeperAPIKey() string
	GetManagerAPIKey() string
	GetMaxMBFile() int
}

type config struct {
	id                string
	managerAPIKey     string
	storekeeperAPIKey string
	dailyCronSchedule string
	maxMBFileUpload   int
}

func NewConfigComponent(id string) *config {
	return &config{id: id}
}

func (c *config) InitFlags() {
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
	flag.StringVar(
		&c.dailyCronSchedule,
		"cron-daily-schedule",
		"0 0 3 * * *", // default: 3:00 AM mỗi ngày
		"6-field cron schedule for daily inventory sync (second minute hour day month weekday)",
	)
	flag.IntVar(
		&c.maxMBFileUpload,
		"max-mb-upload",
		5,
		"max mb csv file (default=5mb)",
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

func (c *config) GetManagerAPIKey() string {
	return c.managerAPIKey
}

func (c *config) GetStorekeeperAPIKey() string {
	return c.storekeeperAPIKey
}

func (c *config) GetDailyCronSchedule() string {
	return c.dailyCronSchedule
}

func (c *config) GetMaxMBFile() int {
	return c.maxMBFileUpload
}
