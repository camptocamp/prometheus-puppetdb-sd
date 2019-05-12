package puppetdb

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

// PuppetDB stores a PuppetDB client
type PuppetDB struct {
	url string

	client *http.Client
}

// NewClient returns a PuppetDB structure
func NewClient(rawURL, certFile, keyFile, caCertFile string, sslSkipVerify bool) (puppetDBClient *PuppetDB, err error) {
	puppetDBClient = &PuppetDB{
		url: rawURL,
	}

	puppetdbURL, err := url.Parse(rawURL)
	if err != nil {
		err = fmt.Errorf("failed to parse URL: %s", err)
		return
	}

	if puppetdbURL.Scheme != "http" && puppetdbURL.Scheme != "https" {
		err = fmt.Errorf("` %s' is not a valid http scheme", puppetdbURL.Scheme)
		return
	}

	var transport *http.Transport
	if puppetdbURL.Scheme == "https" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			err = fmt.Errorf("failed to load client cert: %s", err)
			return nil, err
		}

		// Load CA cert
		caCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			err = fmt.Errorf("failed to load ca cert: %s", err)
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: sslSkipVerify,
		}
		tlsConfig.BuildNameToCertificate()
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	} else {
		transport = &http.Transport{}
	}
	puppetDBClient.client = &http.Client{Transport: transport}
	return
}

// GetTargets requests the PuppetDB to retrieve a list of nodes and
// returns the result as a list of Prometheus static configs
func (p *PuppetDB) GetTargets(query string) (staticConfigs []types.StaticConfig, err error) {
	staticConfigs = []types.StaticConfig{}

	nodes, err := p.getNodes(query)
	if err != nil {
		log.Errorf("failed to get nodes: %s", err)
		return
	}

	for _, node := range nodes {
		for jobName, targets := range node.Exporters {

			// TODO: Remove backward compatibility
			t, err := extractTargets(targets)
			if err != nil {
				log.Errorf("failed to extract targets: %s", err)
				return nil, err
			}

			for _, vt := range t {
				url, err := url.Parse(vt.URL)
				if err != nil {
					return nil, err
				}

				labels := map[string]string{
					"certname":     node.Certname,
					"metrics_path": url.Path,
					"job":          jobName,
					"scheme":       url.Scheme,
				}
				for k, v := range url.Query() {
					labels[fmt.Sprintf("__param_%s", k)] = v[0]
					labels[k] = v[0]
				}
				for k, v := range vt.Labels {
					labels[k] = v
				}
				staticConfig := types.StaticConfig{
					Targets: []string{url.Host},
					Labels:  labels,
				}
				staticConfigs = append(staticConfigs, staticConfig)
			}
		}
	}

	return
}

func (p *PuppetDB) getNodes(query string) (nodes []types.Node, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", query))
	puppetdbURL := fmt.Sprintf("%s/pdb/query/v4", p.url)
	req, err := http.NewRequest("POST", puppetdbURL, form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &nodes)
	return
}

// Allow backward compatibility (to remove)
func extractTargets(targets interface{}) (t []types.Exporter, err error) {
	switch v := targets.(type) {
	case string:
		log.Warningf("Deprecated: target should be a struct Exporter, not a String: %v", v)
		e := types.Exporter{
			URL:    v,
			Labels: make(map[string]string),
		}
		t = []types.Exporter{e}
	case []interface{}:
		switch v[0].(type) {
		case string:
			log.Warningf("Deprecated: target should be a struct Exporter, not an Array of Strings: %v", v)
			t = make([]types.Exporter, len(v))
			for i := range v {
				t[i] = types.Exporter{
					URL:    v[i].(string),
					Labels: make(map[string]string),
				}
			}
		case map[string]interface{}:
			t = make([]types.Exporter, len(v))
			for i := range v {
				a := v[i].(map[string]interface{})
				t[i] = types.Exporter{
					URL:    a["url"].(string),
					Labels: make(map[string]string),
				}
				if a["labels"] != nil {
					for lk, lv := range a["labels"].(map[string]interface{}) {
						t[i].Labels[lk] = lv.(string)
					}
				}
			}
		}
	default:
		err = fmt.Errorf("failed to determine target type: %v", v)
	}
	return
}
