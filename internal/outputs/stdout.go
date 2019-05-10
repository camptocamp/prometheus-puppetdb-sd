package outputs

import (
	"fmt"

	yaml "gopkg.in/yaml.v1"

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

// OutputStdout stores values needed to print output to stdout
type OutputStdout struct{}

// WriteOutput writes static configs to stdout
func (o *OutputStdout) WriteOutput(staticConfigs []types.StaticConfig) (err error) {
	c, err := yaml.Marshal(&staticConfigs)
	if err != nil {
		return
	}
	fmt.Printf("%s", string(c))
	return
}
