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
	cfg, err := config.LoadConfig(version)
	if err != nil {
		return
	}

	o, err := outputs.Setup(&outputs.Options{
		Name:          cfg.Output,
		FilePath:      cfg.File,
		ConfigMapName: cfg.ConfigMap,
		Namespace:     cfg.Namespace,
		ObjectLabels:  cfg.ObjectLabels,
	})
	if err != nil {
		log.Fatalf("failed to setup output: %s", err)
		return
	}

	puppetDBClient, err := puppetdb.NewClient(cfg.PuppetDBURL, cfg.CertFile, cfg.KeyFile, cfg.CACertFile, cfg.SSLSkipVerify)
	if err != nil {
		log.Fatalf("failed to build a PuppetDB client: %s", err)
		return
	}

	for {
		targets, err := puppetDBClient.GetTargets(cfg.Query)
		if err != nil {
			log.Errorf("failed to retrieve exporters: %s", err)
			continue
		}
		err = o.WriteOutput(targets)
		if err != nil {
			log.Errorf("failed to write output: %s", err)
			continue
		}

		log.Infof("Sleeping for %v", cfg.Sleep)
		time.Sleep(cfg.Sleep)
	}
}
