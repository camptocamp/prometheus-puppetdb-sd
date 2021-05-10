package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
)

// Config describes global configuration
type Config struct {
	GeneralConfig `group:"Application Options"`
	PuppetDB      PuppetDBConfig     `group:"PuppetDB Client Options" namespace:"puppetdb"`
	PrometheusSD  PrometheusSDConfig `group:"Prometheus Service Discovery Options" namespace:"prometheus"`
	Output        OutputConfig       `group:"Output Configuration" namespace:"output"`
}

// GeneralConfig describes general application configuration
type GeneralConfig struct {
	Version bool          `short:"V" long:"version" description:"Display version."`
	Manpage bool          `short:"m" long:"manpage" description:"Output manpage."`
	Sleep   time.Duration `short:"s" long:"sleep" description:"Sleep time between queries." env:"SLEEP" default:"5s"`
}

// PuppetDBConfig describes PuppetDB client configuration
type PuppetDBConfig struct {
	URL           string `short:"u" long:"url" description:"PuppetDB base URL." env:"PUPPETDB_URL" default:"http://puppetdb:8080"`
	CertFile      string `short:"x" long:"cert-file" description:"A PEM encoded certificate file." env:"PUPPETDB_CERT_FILE"`
	KeyFile       string `short:"y" long:"key-file" description:"A PEM encoded private key file." env:"PUPPETDB_KEY_FILE"`
	CACertFile    string `short:"z" long:"cacert-file" description:"A PEM encoded CA's certificate file." env:"PUPPETDB_CACERT_FILE"`
	SSLSkipVerify bool   `short:"k" long:"ssl-skip-verify" description:"Skip SSL verification." env:"PUPPETDB_SSL_SKIP_VERIFY"`
	Query         string `short:"q" long:"query" description:"PuppetDB query." env:"PUPPETDB_QUERY" default:"resources[certname, parameters] { type = 'Prometheus::Scrape_job' and exported = true }"`
}

// PrometheusSDConfig describes Prometheus service discovery configuration
type PrometheusSDConfig struct {
	ProxyURL string `long:"proxy-url" description:"Prometheus target scraping proxy URL." env:"PROMETHEUS_PROXY_URL"`
}

// OutputConfig describes output configuration
type OutputConfig struct {
	Method    OutputMethod          `short:"o" long:"method" description:"Output method." choice:"stdout" choice:"file" choice:"k8s-secret" env:"OUTPUT_METHOD" default:"stdout"`
	Format    OutputFormat          `long:"format" description:"Output format." choice:"scrape-configs" choice:"static-configs" choice:"merged-static-configs" env:"OUTPUT_FORMAT" default:"scrape-configs"`
	Stdout    StdoutOutputConfig    `group:"Stdout Output Configuration" namespace:"stdout"`
	File      FileOutputConfig      `group:"File Output Configuration" namespace:"file"`
	K8sSecret K8sSecretOutputConfig `group:"Kubernetes Secret Output Configuration" namespace:"k8s-secret"`
}

// OutputMethod represents an output method
type OutputMethod string

// OutputFormat represents an output format
type OutputFormat string

// StdoutOutputConfig describes stdout output configuration
type StdoutOutputConfig struct{}

// FileOutputConfig describes file output configuration
type FileOutputConfig struct {
	Filename        string `short:"f" long:"filename" description:"Output filename." env:"OUTPUT_FILENAME" default:"puppetdb-sd.yml"`
	FilenamePattern string `long:"filename-pattern" description:"Output filename pattern ('*' is the placeholder)." env:"OUTPUT_FILENAME_PATTERN" default:"*.yml"`
	Directory       string `long:"directory" description:"Output directory." env:"OUTPUT_DIRECTORY" default:"/etc/prometheus/puppetdb-sd"`
}

// K8sSecretOutputConfig describes Kubernetes secret output configuration
type K8sSecretOutputConfig struct {
	SecretName            string            `long:"secret-name" description:"Kubernetes secret name." env:"OUTPUT_K8S_SECRET_NAME"`
	Namespace             string            `long:"namespace" description:"Kubernetes namespace." env:"OUTPUT_K8S_NAMESPACE"`
	ObjectLabels          map[string]string `long:"object-labels" description:"Labels to add to Kubernetes objects." env:"OUTPUT_K8S_OBJECT_LABELS" default:"app.kubernetes.io/name:prometheus-puppetdb-sd"`
	SecretKey             string            `long:"secret-key" description:"Kubernetes secret key." env:"OUTPUT_K8S_SECRET_KEY"`
	SecretKeyPattern      string            `long:"secret-key-pattern" description:"Kubernetes secret key pattern ('*' is the placeholder)." env:"OUTPUT_K8S_SECRET_KEY_PATTERN"`
	ExtraConfigSecretName string            `long:"extra-config-secret-name" description:"Kubernetes secret name containing additional config." env:"OUTPUT_K8S_EXTRA_CONFIG_SECRET_NAME"`
	ExtraConfigSecretKey  string            `long:"extra-config-secret-key" description:"Key of the Kubernetes secret containing additional config." env:"OUTPUT_K8S_EXTRA_CONFIG_SECRET_KEY"`
}

const (
	// Stdout output method prints Prometheus configuration on stdout
	Stdout OutputMethod = "stdout"
	// File output method stores Prometheus configuration into files
	File OutputMethod = "file"
	// K8sSecret output method stores Prometheus configuration into Kubernetes secret
	K8sSecret OutputMethod = "k8s-secret"

	// ScrapeConfigs output format renders a list of Prometheus scrape configurations
	ScrapeConfigs OutputFormat = "scrape-configs"
	// StaticConfigs output format renders a list of Prometheus static configurations per job
	StaticConfigs OutputFormat = "static-configs"
	// MergedStaticConfigs output format renders a unique list of Prometheus scrape configurations for all jobs
	MergedStaticConfigs OutputFormat = "merged-static-configs"
)

// LoadConfig parses arguments
func LoadConfig(version string) (c Config) {
	parser := flags.NewParser(&c, flags.Default)
	args, err := parser.Parse()
	if err != nil {
		os.Exit(2)
	}
	if len(args) != 0 {
		log.Fatalf("Unexpected arguments: %s", args)
	}

	if c.Version {
		fmt.Printf("Prometheus PuppetDB SD v%v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "PuppetDB based service discovery for Prometheus"
		parser.WriteManPage(&buf)
		fmt.Printf("%s", buf.String())
		os.Exit(0)
	}
	return
}
