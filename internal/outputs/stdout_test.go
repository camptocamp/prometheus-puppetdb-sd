package outputs

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/stretchr/testify/assert"
)

func (o *StdoutOutput) testStdoutWriteOutput(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	oldStdout := os.Stdout

	for i := range scrapeConfigs {
		r, w, err := os.Pipe()
		if err != nil {
			assert.FailNow(t, "Failed to redirect stdout to a pipe", err.Error())
		}
		defer r.Close()
		defer w.Close()

		os.Stdout = w

		c := make(chan error)
		go func() {
			err := o.WriteOutput(ctx, scrapeConfigs[i])
			w.Close()
			c <- err
		}()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		if err != nil {
			assert.FailNow(t, "Failed to read output", err.Error())
		}

		err = <-c
		if err != nil {
			assert.FailNow(t, "Failed to write output", err.Error())
		}

		output := buf.String()
		expectedOutput := expectedOutputs[i][o.format].(string)

		assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(output))
	}

	os.Stdout = oldStdout
}

func TestStdoutWriteOutputScrapeConfigsSuccess(t *testing.T) {
	o := StdoutOutput{
		format: config.ScrapeConfigs,
	}

	o.testStdoutWriteOutput(t)
}

func TestStdoutWriteOutputMergedStaticConfigsSuccess(t *testing.T) {
	o := StdoutOutput{
		format: config.MergedStaticConfigs,
	}

	o.testStdoutWriteOutput(t)
}
