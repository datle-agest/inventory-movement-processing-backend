package configc

import sctx "inventory-movement-processing/pkg/service_context"

type config struct {
	id string
}

func NewConfigComponent(id string) *config {
	return &config{id: id}
}

func (c *config) InitFlags() {

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
