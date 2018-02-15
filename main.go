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

	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v1"

	"github.com/jessevdk/go-flags"
)

var version = "undefined"
var transport *http.Transport

type Config struct {
	Version       bool   `short:"V" long:"version" description:"Display version."`
	PuppetDBURL   string `short:"u" long:"puppetdb-url" description:"PuppetDB base URL." env:"PROMETHEUS_PUPPETDB_URL" default:"http://puppetdb:8080"`
	CertFile      string `short:"x" long:"cert-file" description:"A PEM encoded certificate file." env:"PROMETHEUS_CERT_FILE" default:"certs/client.pem"`
	KeyFile       string `short:"y" long:"key-file" description:"A PEM encoded private key file." env:"PROMETHEUS_KEY_FILE" default:"certs/client.key"`
	CACertFile    string `short:"z" long:"cacert-file" description:"A PEM encoded CA's certificate file." env:"PROMETHEUS_CACERT_FILE" default:"certs/cacert.pem"`
	SSLSkipVerify bool   `short:"k" long:"ssl-skip-verify" description:"Skip SSL verification." env:"PROMETHEUS_SSL_SKIP_VERIFY"`
	Query         string `short:"q" long:"puppetdb-query" description:"PuppetDB query." env:"PROMETHEUS_PUPPETDB_QUERY" default:"facts[certname, value] { name='prometheus_exporters' and nodes { deactivated is null } }"`
	ConfigDir     string `short:"c" long:"config-dir" description:"Prometheus config dir." env:"PROMETHEUS_CONFIG_DIR" default:"/etc/prometheus"`
	Sleep         string `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_PUPPETDB_SLEEP" default:"5s"`
	Manpage       bool   `short:"m" long:"manpage" description:"Output manpage."`
}

type Node struct {
	Certname  string            `json:"certname"`
	Exporters map[string]string `json:"value"`
}

type StaticConfig struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}

type ScrapeConfig struct {
	JobName       string         `yaml:"job_name,omitempty"`
	MetricsPath   string         `yaml:"metrics_path,omitempty"`
	Scheme        string         `yaml:"scheme,omitempty"`
	StaticConfigs []StaticConfig `yaml:"static_configs,omitempty"`
}

type PrometheusConfig struct {
	ScrapeConfigs []*ScrapeConfig `yaml:"scrape_configs,omitempty"`
}

func (p *PrometheusConfig) addTarget(jobName, metricsPath, scheme, target, certname string) {
	staticConfig := StaticConfig{
		Targets: []string{target},
		Labels:  map[string]string{"certname": certname},
	}

	for _, config := range p.ScrapeConfigs {
		if config.JobName == jobName && config.MetricsPath == metricsPath && config.Scheme == scheme {
			config.StaticConfigs = append(config.StaticConfigs, staticConfig)
			return
		}
	}

	// Not found
	config := ScrapeConfig{
		JobName:       jobName,
		MetricsPath:   metricsPath,
		Scheme:        scheme,
		StaticConfigs: []StaticConfig{staticConfig},
	}
	p.ScrapeConfigs = append(p.ScrapeConfigs, &config)
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
		fmt.Printf(buf.String())
		os.Exit(0)
	}
	return
}

func getNodes(client *http.Client, puppetdb string, query string) (nodes []Node, err error) {
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

func writeNodes(nodes []Node, dir string) (err error) {
	prometheusConfig := PrometheusConfig{}

	for _, node := range nodes {
		for jobName, target := range node.Exporters {
			url, err := url.Parse(target)
			if err != nil {
				return err
			}
			prometheusConfig.addTarget(jobName, url.Path, url.Scheme, url.Host, node.Certname)
		}
	}
	c, err := yaml.Marshal(&prometheusConfig)
	if err != nil {
		return
	}

	os.MkdirAll(fmt.Sprintf("%s/conf.d", dir), 0755)
	err = ioutil.WriteFile(fmt.Sprintf("%s/conf.d/prometheus-puppetdb.yml", dir), c, 0644)
	if err != nil {
		return
	}

	return nil
}

var client *http.Client
var cfg Config

func init() {
	var err error

	cfg, err = loadConfig(version)
	if err != nil {
		log.Fatal(err)
	}

	puppetdbURL, err := url.Parse(cfg.PuppetDBURL)
	if err != nil {
		log.Fatal(err)
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
	for {
		nodes, err := getNodes(client, cfg.PuppetDBURL, cfg.Query)
		if err != nil {
			log.Error(err)
			break
		}

		err = writeNodes(nodes, cfg.ConfigDir)
		if err != nil {
			log.Error(err)
			break
		}

		sleep, err := time.ParseDuration(cfg.Sleep)
		if err != nil {
			log.Error(err)
			break
		}
		log.Infof("Sleeping for %v", sleep)
		time.Sleep(sleep)
	}
}
