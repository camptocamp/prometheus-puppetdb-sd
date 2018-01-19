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
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

var version = "undefined"
var transport *http.Transport

type Config struct {
	Version         bool   `short:"V" long:"version" description:"Display version."`
	PuppetDBURL     string `short:"u" long:"puppetdb-url" description:"PuppetDB base URL." env:"PROMETHEUS_PUPPETDB_URL" default:"http://puppetdb:8080"`
	CertFile        string `short:"x" long:"cert-file" description:"A PEM encoded certificate file." env:"PROMETHEUS_CERT_FILE" default:"certs/client.pem"`
	KeyFile         string `short:"y" long:"key-file" description:"A PEM encoded private key file." env:"PROMETHEUS_KEY_FILE" default:"certs/client.key"`
	CACertFile      string `short:"z" long:"cacert-file" description:"A PEM encoded CA's certificate file." env:"PROMETHEUS_CACERT_FILE" default:"certs/cacert.pem"`
	SSLSkipVerify   bool   `short:"k" long:"ssl-skip-verify" description:"Skip SSL verification." env:"PROMETHEUS_SSL_SKIP_VERIFY"`
	Query           string `short:"q" long:"puppetdb-query" description:"PuppetDB query." env:"PROMETHEUS_PUPPETDB_QUERY" default:"facts[certname, value]"`
	Filter          string `short:"f" long:"puppetdb-filter" description:"PuppetDB filter." env:"PROMETHEUS_PUPPETDB_FILTER" default:"name='ipaddress' and nodes { deactivated is null }"`
	RoleMappingFile string `short:"r" long:"role-mapping-file" description:"Role mapping configuration file" env:"PROMETHEUS_ROLE_MAPPING_FILE" default:"role-mapping.yaml"`
	TargetsDir      string `short:"c" long:"targets-dir" description:"Directory to store File SD targets files." env:"PROMETHEUS_TARGETS_DIR" default:"/etc/prometheus/targets"`
	Sleep           string `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_PUPPETDB_SLEEP" default:"60s"`
	Manpage         bool   `short:"m" long:"manpage" description:"Output manpage."`
}

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

type RoleMapping struct {
	Exporter string   `yaml:"exporter"`
	Port     int      `yaml:"port"`
	Path     string   `yaml:"path"`
	Scheme   string   `yaml:"scheme"`
	Roles    []string `yaml:"roles"`
}

type Targets struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	//log.SetLevel(log.WarnLevel)
}

func bailout() {
	log.Warn("Caught a signal, bailing out. Bye bye!")
}

func main() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bailout()
		os.Exit(1)
	}()

	cfg, err := loadConfig(version)
	if err != nil {
		log.Fatal("Failed to parse command flags, error=", err)
	}

	log.Info("Prometheus PuppetDB Service Discovery starting!")
	log.Info("prometheus-puppetdb v", version)

	puppetdbURL, err := url.Parse(cfg.PuppetDBURL)
	if err != nil {
		log.Fatal("Couldn't parse PuppetDB URL, error=", err)
	}
	log.Infof("PuppetDB URL for queries: %s", puppetdbURL)

	if puppetdbURL.Scheme != "http" && puppetdbURL.Scheme != "https" {
		log.Fatalf("%s is not a valid scheme for PuppetDB URL (valid options: http or https)", puppetdbURL.Scheme)
	}

	if puppetdbURL.Scheme == "https" {
		log.Info("Setting up https client with Client TLS authentication")
		// Load client cert
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatal("Failed loading client certificate, error=", err)
		}

		// Load CA cert
		caCert, err := ioutil.ReadFile(cfg.CACertFile)
		if err != nil {
			log.Fatal("Failed loading CA's certificate, error=", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		if cfg.SSLSkipVerify {
			log.Warn("Skipping SSL certificate verification!")
		}
		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: cfg.SSLSkipVerify,
		}
		tlsConfig.BuildNameToCertificate()
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	} else {
		log.Info("Setting up http client without Client TLS authentication")
		transport = &http.Transport{}
	}

	// Setup the http client
	client := &http.Client{Transport: transport}

	log.Info("Starting service discovery loop")
	// Start the main loop
	for {
		// Read the role mapping from configuration file
		roleMapping, err := loadRoleMapping(cfg.RoleMappingFile)
		if err != nil {
			log.Fatal("Couldn't load Role Mapping configuration file, error=", err)
		}

		// Clean the targets directory, remove any target files that are no longer listed in Role Mapping
		err = cleanupTargetsDir(cfg.TargetsDir, roleMapping)
		if err != nil {
			log.Error("Error cleaning up targets directory, error=", err)
		}

		// Iterate through the Exporters
		for e := range roleMapping {
			log.Infof("Starting discovery for job=%s", roleMapping[e].Exporter)
			var nodes []Node
			// Iterate through the Roles mapped to each Exporter
			for r := range roleMapping[e].Roles {
				log.Infof("Collecting nodes information for job=%s role=%s", roleMapping[e].Exporter, roleMapping[e].Roles[r])
				var tmpNodes []Node
				// Get the nodes for this role
				tmpNodes, err = getNodes(client, cfg.PuppetDBURL, cfg.Query, cfg.Filter, roleMapping[e].Roles[r])
				if err != nil {
					log.Error("Failed to fetch nodes from PuppetDB, error=", err)
					break
				}
				nodes = append(nodes, tmpNodes...)
			}

			if err == nil {
				log.Infof("Writing nodes information to target file for job=%s", roleMapping[e].Exporter)
				// Write the nodes to a Targets file per Exporter (==Job)
				err = writeNodes(nodes, roleMapping[e].Port, roleMapping[e].Path, roleMapping[e].Scheme, roleMapping[e].Exporter, cfg.TargetsDir)
				if err != nil {
					log.Error("Couldn't write target file, error=", err)
					break
				}
			} else {
				log.Warn("Node collection failed, not updating targets")
			}
		}

		// Sleep...
		sleep, err := time.ParseDuration(cfg.Sleep)
		if err != nil {
			log.Error("Failed to parse sleep duration, falling back to 60s, error=", err)
			sleep = time.Minute
		}
		log.Infof("Sleeping for %v", sleep)
		time.Sleep(sleep)
		log.Info("Wake up and start again...")
	}
}

func loadConfig(version string) (c Config, err error) {
	parser := flags.NewParser(&c, flags.Default)
	_, err = parser.Parse()
	if err != nil {
		return
	}

	if c.Version {
		fmt.Printf("Prometheus-puppetdb v%v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "Prometheus service discovery based on PuppetDB"
		parser.WriteManPage(&buf)
		fmt.Printf(buf.String())
		os.Exit(0)
	}
	return
}

func loadRoleMapping(mappingFile string) (roleMapping []RoleMapping, err error) {
	filename, _ := filepath.Abs(mappingFile)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlFile, &roleMapping)
	if err != nil {
		return
	}

	log.Info("Role mapping configuration loaded")
	return
}

// Iterate through the yml & yaml files in TargetsDir and remove all that do not match an Exporter in roleMapping
func cleanupTargetsDir(dir string, roles []RoleMapping) (err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

OUTER:
	for _, file := range files {
		for r := range roles {
			found, _ := regexp.MatchString(fmt.Sprintf("%s.(yaml|yml)", roles[r].Exporter), file.Name())
			if found {
				continue OUTER
			}
		}

		err = os.Remove(fmt.Sprintf("%s/%s", dir, file.Name()))
		if err != nil {
			return
		}
	}
	return
}

func getNodes(client *http.Client, puppetdb string, query string, filter string, role string) (nodes []Node, err error) {
	// Build the query from Query, Filter and the role
	q := fmt.Sprintf("%s { %s and facts { name='role' and value='%s' } }", query, filter, role)

	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", q))
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

func writeNodes(nodes []Node, port int, path string, scheme string, job string, dir string) (err error) {
	allTargets := []Targets{}

	for _, node := range nodes {
		targets := Targets{}

		target := fmt.Sprintf("%s:%v", node.Ipaddress, port)
		targets.Targets = append(targets.Targets, target)
		targets.Labels = map[string]string{
			"job":          job,
			"certname":     node.Certname,
			"metrics_path": path,
			"scheme":       scheme,
		}
		allTargets = append(allTargets, targets)
	}

	d, err := yaml.Marshal(&allTargets)
	if err != nil {
		return
	}

	os.MkdirAll(fmt.Sprintf("%s", dir), 0755)
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.yml", dir, job), d, 0644)
	if err != nil {
		return
	}

	return nil
}
