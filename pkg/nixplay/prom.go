package nixplay

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type promImpl struct {
	promFactory promauto.Factory
}

func (c *clientImpl) promRegister(reg prometheus.Registerer) error {
	c.prom.promFactory = promauto.With(reg)

	return nil
}
