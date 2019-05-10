package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"github.com/camptocamp/prometheus-puppetdb/internal/outputs"
	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

var version = "undefined"
var transport *http.Transport

// Config stores handler's configuration
type Config struct {
	Version       bool          `short:"V" long:"version" description:"Display version."`
	PuppetDBURL   string        `short:"u" long:"puppetdb-url" description:"PuppetDB base URL." env:"PROMETHEUS_PUPPETDB_URL" default:"http://puppetdb:8080"`
	CertFile      string        `short:"x" long:"cert-file" description:"A PEM encoded certificate file." env:"PROMETHEUS_CERT_FILE" default:"certs/client.pem"`
	KeyFile       string        `short:"y" long:"key-file" description:"A PEM encoded private key file." env:"PROMETHEUS_KEY_FILE" default:"certs/client.key"`
	CACertFile    string        `short:"z" long:"cacert-file" description:"A PEM encoded CA's certificate file." env:"PROMETHEUS_CACERT_FILE" default:"certs/cacert.pem"`
	SSLSkipVerify bool          `short:"k" long:"ssl-skip-verify" description:"Skip SSL verification." env:"PROMETHEUS_SSL_SKIP_VERIFY"`
	Query         string        `short:"q" long:"puppetdb-query" description:"PuppetDB query." env:"PROMETHEUS_PUPPETDB_QUERY" default:"facts[certname, value] { name='prometheus_exporters' and nodes { deactivated is null } }"`
	Output        string        `short:"o" long:"output" description:"Output. One of stdout, file or configmap" env:"PROMETHEUS_PUPPETDB_OUTPUT" default:"stdout"`
	File          string        `short:"f" long:"config-file" description:"Prometheus target file." env:"PROMETHEUS_PUPPETDB_FILE" default:"/etc/prometheus/targets/prometheus-puppetdb/targets.yml"`
	ConfigMap     string        `long:"configmap" description:"Kubernetes ConfigMap to update." env:"PROMETHEUS_PUPPETDB_CONFIGMAP" default:"prometheus-puppetdb"`
	NameSpace     string        `long:"namespace" description:"Kubernetes NameSpace to use." env:"PROMETHEUS_PUPPETDB_NAMESPACE" default:"default"`
	Sleep         time.Duration `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_PUPPETDB_SLEEP" default:"5s"`
	Manpage       bool          `short:"m" long:"manpage" description:"Output manpage."`
}

func loadConfig(version string) (c Config, err error) {
	parser := flags.NewParser(&c, flags.Default)
	_, err = parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	if c.Version {
		fmt.Printf("Prometheus-puppetdb v%v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "Prometheus scrape lists based on PuppetDB"
		parser.WriteManPage(&buf)
		fmt.Printf("%s", buf.String())
		os.Exit(0)
	}
	return
}

func getNodes(client *http.Client, puppetdb string, query string) (nodes []types.Node, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", query))
	puppetdbURL := fmt.Sprintf("%s/pdb/query/v4", puppetdb)
	req, err := http.NewRequest("POST", puppetdbURL, form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
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

func getTargets() (staticConfigs []types.StaticConfig, err error) {
	staticConfigs = []types.StaticConfig{}

	nodes, err := getNodes(client, cfg.PuppetDBURL, cfg.Query)
	if err != nil {
		log.Errorf("failed to get nodes: %v", err)
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
	/*
		c, err = yaml.Marshal(&fileSdConfig)
		if err != nil {
			return
		}
	*/

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

var client *http.Client
var cfg Config

func init() {
	var err error

	cfg, err = loadConfig(version)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	puppetdbURL, err := url.Parse(cfg.PuppetDBURL)
	if err != nil {
		log.Fatalf("failed to parse PuppetDB URL: %v", err)
	}

	if puppetdbURL.Scheme != "http" && puppetdbURL.Scheme != "https" {
		log.Fatalf("%s is not a valid http scheme\n", puppetdbURL.Scheme)
	}

	if puppetdbURL.Scheme == "https" {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatal(err)
		}

		// Load CA cert
		caCert, err := ioutil.ReadFile(cfg.CACertFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: cfg.SSLSkipVerify,
		}
		tlsConfig.BuildNameToCertificate()
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	} else {
		transport = &http.Transport{}
	}

	client = &http.Client{Transport: transport}
}

func main() {
	o, err := outputs.Setup(&outputs.Options{
		Name:     cfg.Output,
		FilePath: cfg.File,
	})
	if err != nil {
		log.Fatalf("failed to setup output: %s", err)
		return
	}

	for {
		targets, err := getTargets()
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
