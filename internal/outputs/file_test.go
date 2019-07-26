package outputs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/camptocamp/prometheus-puppetdb/internal/config"
)

func TestFileSetupSuccess(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "prometheus-puppetdb-test")
	if err != nil {
		assert.FailNow(t, "Failed to create temporary directory", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	dir := tmpDir + "/etc/prometheus"

	cfg := config.OutputConfig{
		File: config.FileOutputConfig{
			Directory: dir,
		},
	}

	_, err = setupFileOutput(&cfg)

	assert.Nil(t, err)
	assert.DirExists(t, dir)
}

func (o *FileOutput) testFileWriteOutput(t *testing.T) {
	directory, err := ioutil.TempDir("", "prometheus-puppetdb-test")
	if err != nil {
		assert.FailNow(t, "Failed to create temporary directory", err.Error())
	}
	defer os.RemoveAll(directory)

	o.filename = "puppetdb.yml"
	o.filenamePattern = "puppetdb-*.yml"
	o.directory = directory

	oldPaths := map[string]bool{}

	for i := range scrapeConfigs {
		err = o.WriteOutput(scrapeConfigs[i])
		if err != nil {
			assert.FailNow(t, "Failed to write output", err.Error())
		}

		switch o.format {
		case config.ScrapeConfigs, config.MergedStaticConfigs:
			path := fmt.Sprintf("%s/%s", o.directory, o.filename)

			buf, err := ioutil.ReadFile(path)
			if err != nil {
				assert.FailNow(t, "Failed to read output file content", err.Error())
			}

			output := string(buf)
			expectedOutput := expectedOutputs[i][o.format].(string)

			assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(output))
		case config.StaticConfigs:
			paths := map[string]bool{}

			for _, scrapeConfig := range scrapeConfigs[i] {
				jobName := scrapeConfig.JobName

				path := strings.Replace(fmt.Sprintf("%s/%s", o.directory, o.filenamePattern), "*", jobName, 1)

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					assert.FailNow(t, "Failed to read output file content", err.Error())
				}

				output := string(buf)
				expectedOutput := expectedOutputs[i][o.format].(map[string]string)[jobName]

				assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(output))

				paths[path] = true
				delete(oldPaths, path)
			}

			for path := range oldPaths {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					filename := filepath.Base(path)

					assert.Fail(t, "Unexpected filename in output directory", "filename: %s", filename)
				}
			}

			oldPaths = paths
		}
	}
}

func TestFileWriteOutputScrapeConfigsSuccess(t *testing.T) {
	o := FileOutput{
		format: config.ScrapeConfigs,
	}

	o.testFileWriteOutput(t)
}

func TestFileWriteOutputStaticConfigsSuccess(t *testing.T) {
	o := FileOutput{
		format: config.StaticConfigs,
	}

	o.testFileWriteOutput(t)
}

func TestFileWriteOutputMergedStaticConfigsSuccess(t *testing.T) {
	o := FileOutput{
		format: config.MergedStaticConfigs,
	}

	o.testFileWriteOutput(t)
}
