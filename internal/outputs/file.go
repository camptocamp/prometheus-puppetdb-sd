package outputs

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v1"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/types"
)

// FileOutput stores values needed by the File output
type FileOutput struct {
	filename        string
	filenamePattern string
	directory       string

	format config.OutputFormat

	state struct {
		oldPaths map[string]bool
	}
}

func setupFileOutput(cfg *config.OutputConfig) (*FileOutput, error) {
	err := os.MkdirAll(cfg.File.Directory, 0755)
	return &FileOutput{
		filename:        cfg.File.Filename,
		filenamePattern: cfg.File.FilenamePattern,
		directory:       cfg.File.Directory,

		format: cfg.Format,
	}, err
}

// WriteOutput writes Prometheus configuration to files
func (o *FileOutput) WriteOutput(scrapeConfigs []*types.ScrapeConfig) (err error) {
	var c []byte
	var mc []byte

	switch o.format {
	case config.ScrapeConfigs:
		c, err = yaml.Marshal(scrapeConfigs)
		if err != nil {
			return
		}

		path := fmt.Sprintf("%s/%s", o.directory, o.filename)

		err = writeFile(path, c)
		if err != nil {
			return
		}
	case config.StaticConfigs, config.MergedStaticConfigs:
		paths := map[string]bool{}

		for _, scrapeConfig := range scrapeConfigs {
			c, err = yaml.Marshal(scrapeConfig.StaticConfigs)
			if err != nil {
				return
			}

			if o.format == config.MergedStaticConfigs {
				mc = append(mc, c...)
			} else {
				path := fmt.Sprintf("%s/%s", o.directory, strings.Replace(o.filenamePattern, "*", scrapeConfig.JobName, 1))

				err = writeFile(path, c)
				if err != nil {
					return
				}

				paths[path] = true
				delete(o.state.oldPaths, path)
			}
		}

		if o.format == config.MergedStaticConfigs {
			path := fmt.Sprintf("%s/%s", o.directory, o.filename)

			err = writeFile(path, mc)
			if err != nil {
				return
			}
		} else {
			for path := range o.state.oldPaths {
				err = os.Remove(path)
				if err != nil {
					return
				}
			}
		}

		o.state.oldPaths = paths
	default:
		err = fmt.Errorf("unexpected output format '%s'", o.format)

		return
	}

	return
}

func writeFile(path string, content []byte) (err error) {
	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(tmpPath)
		}
	}()

	_, err = f.Write(content)
	if err != nil {
		return
	}

	err = f.Sync()
	if err != nil {
		return
	}

	err = f.Close()
	if err != nil {
		return
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		return
	}

	return
}
