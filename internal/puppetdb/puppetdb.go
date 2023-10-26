package puppetdb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/types"
)

// PuppetDB stores a PuppetDB client
type PuppetDB struct {
	client *http.Client

	url   string
	query string
}

// NewClient returns a PuppetDB structure
func NewClient(cfg *config.PuppetDBConfig) (puppetDBClient *PuppetDB, err error) {
	puppetDBClient = &PuppetDB{
		url:   cfg.URL,
		query: cfg.Query,
	}

	puppetdbURL, err := url.Parse(cfg.URL)
	if err != nil {
		err = fmt.Errorf("failed to parse URL: %s", err)
		return
	}

	if puppetdbURL.Scheme != "http" && puppetdbURL.Scheme != "https" {
		err = fmt.Errorf("'%s' is not a valid http scheme", puppetdbURL.Scheme)
		return
	}

	var transport = &http.Transport{}
	var tlsConfig *tls.Config
	if puppetdbURL.Scheme == "https" {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: cfg.SSLSkipVerify,
		}
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	// Assume SSL authentication is required
	if cfg.CertFile != "" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			err = fmt.Errorf("failed to load client cert: %s", err)
			return nil, err
		}

		// Load CA cert
		caCert, err := os.ReadFile(cfg.CACertFile)
		if err != nil {
			err = fmt.Errorf("failed to load ca cert: %s", err)
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: cfg.SSLSkipVerify,
		}
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	puppetDBClient.client = &http.Client{Transport: transport}
	return
}

// GetScrapeConfigs requests the PuppetDB to retrieve a list of nodes with their
// associated endpoints and returns a list of Prometheus scrape configurations
func (p *PuppetDB) GetScrapeConfigs(cfg *config.PrometheusSDConfig) (scrapeConfigs []*types.ScrapeConfig, err error) {
	scrapeConfigs = []*types.ScrapeConfig{}
	scrapeConfigMap := map[string]*types.ScrapeConfig{}

	resources, err := p.getResources()
	if err != nil {
		err = fmt.Errorf("failed to get resources: %s", err)
		return
	}

	for _, resource := range resources {
		certname := resource.Certname
		parameters := resource.Parameters

		jobName := parameters.JobName
		targets := parameters.Targets
		labels := parameters.Labels

		if targets == nil {
			continue
		}

		if labels == nil {
			labels = map[string]string{}
		}

		scrapeConfig, ok := scrapeConfigMap[jobName]
		if !ok {
			scrapeConfig = &types.ScrapeConfig{
				JobName:  jobName,
				ProxyURL: cfg.ProxyURL,
			}
			scrapeConfigs = append(scrapeConfigs, scrapeConfig)
			scrapeConfigMap[jobName] = scrapeConfig
		}

		staticConfigs := &scrapeConfig.StaticConfigs

		if scrapeConfig.ProxyURL != "" {
			if scheme, ok := labels["__scheme__"]; ok {
				labels["__scheme__"] = "http"
				labels["__param__scheme"] = scheme
			}
		}

		labels["certname"] = certname

		staticConfig := &types.StaticConfig{
			Targets: targets,
			Labels:  labels,
		}

		*staticConfigs = append(*staticConfigs, staticConfig)
	}

	return
}

func (p *PuppetDB) getResources() (resources []*types.Resource, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", p.query))
	puppetdbURL := fmt.Sprintf("%s/pdb/query/v4", p.url)
	req, err := http.NewRequest("POST", puppetdbURL, form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		err = fmt.Errorf("HTTP request failed (%s)", err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read HTTP response body (%s)", err)
		return
	}

	err = json.Unmarshal(body, &resources)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal HTTP response body to JSON (%s)", err)
	}
	return
}
