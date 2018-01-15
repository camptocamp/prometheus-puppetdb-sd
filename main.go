package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"gopkg.in/yaml.v2"
)

var version = "undefined"

type Config struct {
	Version       bool   `short:"V" long:"version" description:"Display version."`
	PuppetDBURL   string `short:"u" long:"puppetdb-url" description:"PuppetDB base URL." env:"PROMETHEUS_PUPPETDB_URL" default:"http://puppetdb:8080"`
	CertFile      string `short:"x" long:"cert-file" description:"A PEM eoncoded certificate file." env:"PROMETHEUS_CERT_FILE" default:"certs/client.pem"`
	KeyFile       string `short:"y" long:"key-file" description:"A PEM eoncoded private key file." env:"PROMETHEUS_KEY_FILE" default:"certs/client.key"`
	CACertFile    string `short:"z" long:"cacert-file" description:"A PEM eoncoded CA's certificate file." env:"PROMETHEUS_CACERT_FILE" default:"certs/cacert.pem"`
	SSLSkipVerify bool   `short:"k" long:"ssl-skip-verify" description:"Skip SSL verification." env:"PROMETHEUS_SSL_SKIP_VERIFY"`
	Query         string `short:"q" long:"puppetdb-query" description:"PuppetDB query." env:"PROMETHEUS_PUPPETDB_QUERY" default:"facts[certname, value] { name='ipaddress' and nodes { deactivated is null } }"`
	Port          int    `short:"p" long:"collectd-port" description:"Collectd port." env:"PROMETHEUS_PUPPETDB_COLLECTD_PORT" default:"9103"`
	ConfigDir     string `short:"c" long:"config-dir" description:"Prometheus config dir." env:"PROMETHEUS_CONFIG_DIR" default:"/etc/prometheus"`
	File          string `short:"f" long:"config-file" description:"Prometheus target file." env:"PROMETHEUS_PUPPETDB_FILE" default:"/etc/prometheus/targets/prometheus-puppetdb/targets.yml"`
	Sleep         string `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_PUPPETDB_SLEEP" default:"5s"`
	Manpage       bool   `short:"m" long:"manpage" description:"Output manpage."`
}

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

type Override struct {
	Certname string                 `json:"certname"`
	Override map[string]interface{} `json:"value"`
}

type Targets struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}

type FileSdConfig struct {
	Files []string `yaml:"files,omitempty"`
}

type ScrapeConfig struct {
	JobName       string         `yaml:"job_name,omitempty"`
	MetricsPath   string         `yaml:"metrics_path,omitempty"`
	Scheme        string         `yaml:"scheme,omitempty"`
	FileSdConfigs []FileSdConfig `yaml:"file_sd_configs,omitempty"`
}

type PrometheusConfig struct {
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs,omitempty"`
}

func main() {
	cfg, err := loadConfig(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(cfg.CACertFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	for {
		nodes, err := getNodes(client, cfg.PuppetDBURL, cfg.Query)
		if err != nil {
			fmt.Println(err)
			break
		}

		overrides, err := getOverrides(client, cfg.PuppetDBURL)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = writeNodes(nodes, overrides, cfg.Port, cfg.ConfigDir, cfg.File)
		if err != nil {
			fmt.Println(err)
			break
		}

		sleep, err := time.ParseDuration(cfg.Sleep)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("Sleeping for %v\n", sleep)
		time.Sleep(sleep)
	}
}

func loadConfig(version string) (c Config, err error) {
	parser := flags.NewParser(&c, flags.Default)
	_, err = parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

func getOverrides(client *http.Client, puppetdb string) (overrides map[string]map[string]interface{}, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", "facts[certname, value] { name='prometheus_target_conf' }"))
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

	var nodes []Override
	err = json.Unmarshal(body, &nodes)

	overrides = make(map[string]map[string]interface{})
	for _, node := range nodes {
		overrides[node.Certname] = node.Override
	}

	return
}

func writeNodes(nodes []Node, overrides map[string]map[string]interface{}, port int, dir string, file string) (err error) {
	allTargets := []Targets{}

	prometheusConfig := PrometheusConfig{}

	prometheusConfig.ScrapeConfigs = append(
		prometheusConfig.ScrapeConfigs,
		ScrapeConfig{
			JobName: "prometheus-puppetdb",
			FileSdConfigs: []FileSdConfig{
				{Files: []string{fmt.Sprintf("%s/targets/prometheus-puppetdb/*.yml", dir)}},
			},
		})

	for _, node := range nodes {
		var targets = Targets{}
		var hostname = node.Ipaddress
		var zeport = port
		if o, ok := overrides[node.Certname]; ok {
			if h, ok := o["hostname"]; ok {
				hostname = h.(string)
			}
			if p, ok := o["port"]; ok {
				zeport = int(p.(float64))
			}
			scheme, okScheme := o["scheme"]
			metricsPath, okMetricsPath := o["metrics_path"]
			if okScheme || okMetricsPath {
				var scrapeConfig = ScrapeConfig{}
				scrapeConfig.JobName = node.Certname
				if okScheme {
					scrapeConfig.Scheme = scheme.(string)
				}
				if okMetricsPath {
					scrapeConfig.MetricsPath = metricsPath.(string)
				}
				scrapeConfig.FileSdConfigs = []FileSdConfig{
					{Files: []string{fmt.Sprintf("%s/targets/%s/*.yml", dir, node.Certname)}},
				}
				prometheusConfig.ScrapeConfigs = append(prometheusConfig.ScrapeConfigs, scrapeConfig)

				var target = fmt.Sprintf("%s:%v", hostname, zeport)
				targets.Targets = append(targets.Targets, target)
				targets.Labels = map[string]string{
					"job":      "collectd",
					"certname": node.Certname,
					"host":     node.Certname,
				}

				d, err := yaml.Marshal([]Targets{targets})
				if err != nil {
					return err
				}

				os.MkdirAll(fmt.Sprintf("%s/targets/%s/", dir, node.Certname), 0755)
				err = ioutil.WriteFile(fmt.Sprintf("%s/targets/%s/%s.yml", dir, node.Certname, node.Certname), d, 0644)
				if err != nil {
					return err
				}
			} else {
				var target = fmt.Sprintf("%s:%v", hostname, zeport)
				targets.Targets = append(targets.Targets, target)
				targets.Labels = map[string]string{
					"job":      "collectd",
					"certname": node.Certname,
					"host":     node.Certname,
				}
				allTargets = append(allTargets, targets)
			}
		} else {
			var target = fmt.Sprintf("%s:%v", hostname, zeport)
			targets.Targets = append(targets.Targets, target)
			targets.Labels = map[string]string{
				"job":      "collectd",
				"certname": node.Certname,
				"host":     node.Certname,
			}
			allTargets = append(allTargets, targets)
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

	d, err := yaml.Marshal(&allTargets)
	if err != nil {
		return
	}

	os.MkdirAll(fmt.Sprintf("%s/targets/prometheus-puppetdb/", dir), 0755)
	err = ioutil.WriteFile(file, d, 0644)
	if err != nil {
		return
	}

	return nil
}
