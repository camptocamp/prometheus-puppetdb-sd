package outputs

import (
	"fmt"

	yaml "gopkg.in/yaml.v1"
)

// OutputStdout stores values needed to print output to stdout
type OutputStdout struct{}

// WriteOutput writes data to stdout
func (o *OutputStdout) WriteOutput(data interface{}) (err error) {
	c, err := yaml.Marshal(&data)
	if err != nil {
		return
	}
	fmt.Printf("%s", string(c))
	return
}
