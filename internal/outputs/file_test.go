package outputs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSetupSuccess(t *testing.T) {
	dir, err := ioutil.TempDir("", "testing-prometheus-puppetdb")
	if err != nil {
		assert.FailNow(t, "failed to create tmp dir", err.Error())
	}
	os.RemoveAll(dir)

	_, err = setupOutputFile(filepath.Join(dir, "output.yaml"))
	defer os.RemoveAll(dir)

	assert.Nil(t, err)
	assert.DirExists(t, dir)
}

func TestFileWriteOutputSuccess(t *testing.T) {
	dir, err := ioutil.TempDir("", "testing-prometheus-puppetdb")
	if err != nil {
		assert.FailNow(t, "failed to create tmp dir", err.Error())
	}
	defer os.RemoveAll(dir)
	tmpfn := filepath.Join(dir, "output.yaml")

	data := map[string]string{
		"foo": "bar",
	}
	o := &OutputFile{
		path: tmpfn,
	}

	err = o.WriteOutput(data)

	assert.Nil(t, err)
	fileContent, err := ioutil.ReadFile(tmpfn)
	if err != nil {
		assert.FailNow(t, "failed to read output file content", err.Error())
	}
	assert.Equal(t, "foo: bar\n", string(fileContent))
}
