package outputs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v1"

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

// OutputFile stores values needed by the File output
type OutputFile struct {
	path string
}

func setupOutputFile(path string) (*OutputFile, error) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	return &OutputFile{
		path: path,
	}, err
}

// WriteOutput writes static configs to a file
func (o *OutputFile) WriteOutput(staticConfigs []types.StaticConfig) (err error) {
	os.MkdirAll(filepath.Dir(o.path), 0755)
	c, err := yaml.Marshal(&staticConfigs)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(o.path, c, 0644)
	return
}
