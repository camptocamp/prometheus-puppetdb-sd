package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/outputs"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/puppetdb"
)

var version = "undefined"

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig(version)

	o, err := outputs.Setup(&cfg.Output)
	if err != nil {
		log.Fatalf("Failed to setup output: %s", err)
		return
	}

	puppetDBClient, err := puppetdb.NewClient(&cfg.PuppetDB)
	if err != nil {
		log.Fatalf("Failed to build a PuppetDB client: %s", err)
		return
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for ctx.Err() == nil {
			scrapeConfigs, err := puppetDBClient.GetScrapeConfigs(ctx, &cfg.PrometheusSD)
			if err != nil {
				log.Errorf("Failed to generate scrape_configs: %s", err)
			} else {
				err = o.Write(ctx, scrapeConfigs)
				if err != nil {
					log.Errorf("Failed to write output: %s", err)
				}
			}

			log.Infof("Sleeping for %v", cfg.Sleep)

			select {
			case <-time.After(cfg.Sleep):
			case <-ctx.Done():
			}
		}

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := o.Close(ctx)
		if err != nil {
			log.Errorf("Failed to close output: %s", err)
		}
	}()

	<-sigChan
}
