package outputs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
)

var expectedHttpOutputs = []string{
	`
[
	{
		"targets": [
			"server-1.example.com:9100"
		],
		"labels": {
			"certname": "server-1.example.com",
			"environment": "production",
			"team": "team-1"
		}
	},
	{
		"targets": [
			"server-2.example.com:9100"
		],
		"labels": {
			"certname": "server-2.example.com",
			"environment": "development",
			"team": "team-1"
		}
	},
	{
		"targets": [
			"server-1.example.com:9117"
		],
		"labels": {
			"certname": "server-1.example.com",
			"environment": "production",
			"team": "team-2"
		}
	}
]
`,
	`
[
	{
		"targets": [
			"server-1.example.com:9100"
		],
		"labels": {
			"certname": "server-1.example.com",
			"environment": "production",
			"team": "team-1"
		}
	},
	{
		"targets": [
			"server-2.example.com:9100"
		],
		"labels": {
			"certname": "server-2.example.com",
			"environment": "development",
			"team": "team-2"
		}
	}
]
	`,
}

func TestHttpOutput(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cfg := config.OutputConfig{
		Http: config.HttpOutputConfig{
			Port: 8000,
		},
	}

	o, err := setupHttpOutput(&cfg)

	url := fmt.Sprintf("http://localhost:%d", cfg.Http.Port)

	assert.Nil(t, err)

	resp, err := http.Get(url)
	if err != nil {
		assert.Fail(t, "Failed to request HTTP server", err)
	}

	assert.Equal(t, 404, resp.StatusCode)

	for i := range scrapeConfigs {
		err := o.Write(ctx, scrapeConfigs[i])
		if err != nil {
			assert.FailNow(t, "Failed to write output", err.Error())
		}

		resp, err := http.Get(url)
		if err != nil {
			assert.Fail(t, "Failed to request HTTP server", err)
		}

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		output, err := io.ReadAll(resp.Body)
		if err != nil {
			assert.Fail(t, "Failed to read HTTP response", err)
		}

		expectedOutput := expectedHttpOutputs[i]

		assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(string(output)))
	}

	err = o.Close(ctx)
	if err != nil {
		assert.FailNow(t, "Failed to close output", err.Error())
	}

	_, err = http.Get(url)
	assert.True(t, errors.Is(err, syscall.ECONNREFUSED))
}
