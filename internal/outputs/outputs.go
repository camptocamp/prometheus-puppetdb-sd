package outputs

import (
	"fmt"

	"github.com/camptocamp/prometheus-puppetdb/internal/config"
	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

// Output is an abstraction to the different output types
type Output interface {
	WriteOutput(scrapeConfigs []*types.ScrapeConfig) (err error)
}

// Setup returns an output type
func Setup(cfg *config.OutputConfig) (Output, error) {
	switch cfg.Method {
	case config.Stdout:
		return setupStdoutOutput(cfg)
	case config.File:
		return setupFileOutput(cfg)
	case config.K8sSecret:
		return setupK8sSecretOutput(cfg)
	default:
		return nil, fmt.Errorf("unknown output method: '%s'", cfg.Method)
	}
}
