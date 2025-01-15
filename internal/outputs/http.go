package outputs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/types"
)

// HttpOutput stores values needed by the HTTP output
type HttpOutput struct {
	server *http.Server
	output atomic.Pointer[bytes.Buffer]
}

func setupHttpOutput(cfg *config.OutputConfig) (*HttpOutput, error) {
	o := &HttpOutput{
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Http.Port),
		},
	}

	http.Handle("/", o)

	go func() {
		if err := o.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start HTTP server: %s", err)
		}
	}()

	return o, nil
}

// Write serves Prometheus configuration through an HTTP server
func (o *HttpOutput) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if o.output.Load() == nil {
		w.WriteHeader(404)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	_, err := io.Copy(w, o.output.Load())
	if err != nil {
		log.Errorf("Failed to write response: %s", err)
	}
}

// Write formats Prometheus configuration to be served through HTTP
func (o *HttpOutput) Write(ctx context.Context, scrapeConfigs []*types.ScrapeConfig) (err error) {
	var staticConfigs []*types.StaticConfig

	for _, scrapeConfig := range scrapeConfigs {
		staticConfigs = append(staticConfigs, scrapeConfig.StaticConfigs...)
	}

	var b bytes.Buffer

	encoder := json.NewEncoder(&b)
	encoder.SetIndent("", "\t")

	err = encoder.Encode(staticConfigs)
	if err != nil {
		return
	}

	o.output.Store(&b)

	return
}

// Close shuts down the HTTP server serving Prometheus configuration
func (o *HttpOutput) Close(ctx context.Context) (err error) {
	return o.server.Shutdown(ctx)
}
