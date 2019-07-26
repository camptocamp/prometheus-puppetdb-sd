package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/camptocamp/prometheus-puppetdb/internal/config"
	"github.com/camptocamp/prometheus-puppetdb/internal/outputs"
	"github.com/camptocamp/prometheus-puppetdb/internal/puppetdb"
)

var version = "undefined"

func main() {
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

	for {
		scrapeConfigs, err := puppetDBClient.GetScrapeConfigs(&cfg.PrometheusSD)
		if err != nil {
			log.Errorf("Failed to generate scrape_configs: %s", err)
		} else {
			err = o.WriteOutput(scrapeConfigs)
			if err != nil {
				log.Errorf("Failed to write output: %s", err)
			}
		}

		log.Infof("Sleeping for %v", cfg.Sleep)
		time.Sleep(cfg.Sleep)
	}
}
