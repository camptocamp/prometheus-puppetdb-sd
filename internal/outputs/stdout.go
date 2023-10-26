package outputs

import (
	"context"
	"fmt"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/types"
	yaml "gopkg.in/yaml.v1"
)

// StdoutOutput stores values needed to print output to stdout
type StdoutOutput struct {
	format config.OutputFormat
}

func setupStdoutOutput(cfg *config.OutputConfig) (*StdoutOutput, error) {
	return &StdoutOutput{
		format: cfg.Format,
	}, nil
}

// WriteOutput writes Prometheus configuration to stdout
func (o *StdoutOutput) WriteOutput(ctx context.Context, scrapeConfigs []*types.ScrapeConfig) (err error) {
	var c []byte

	switch o.format {
	case config.ScrapeConfigs:
		c, err = yaml.Marshal(scrapeConfigs)
		if err != nil {
			return
		}

		fmt.Printf("%s", string(c))
	case config.MergedStaticConfigs:
		for _, scrapeConfig := range scrapeConfigs {
			c, err = yaml.Marshal(scrapeConfig.StaticConfigs)
			if err != nil {
				return
			}

			fmt.Printf("%s", string(c))
		}
	default:
		err = fmt.Errorf("unexpected output format '%s'", o.format)

		return
	}

	return
}
