package outputs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v1"
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

// WriteOutput writes data to a file
func (o *OutputFile) WriteOutput(data interface{}) (err error) {
	os.MkdirAll(filepath.Dir(o.path), 0755)
	c, err := yaml.Marshal(&data)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(o.path, c, 0644)
	return
}
