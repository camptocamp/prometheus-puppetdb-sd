package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
)

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
	Namespace     string        `long:"namespace" description:"Kubernetes NameSpace to use." env:"PROMETHEUS_PUPPETDB_NAMESPACE"`
	Sleep         time.Duration `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_PUPPETDB_SLEEP" default:"5s"`
	Manpage       bool          `short:"m" long:"manpage" description:"Output manpage."`
}

// LoadConfig parses arguments and returns a Config structure
func LoadConfig(version string) (c Config, err error) {
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
		parser.ShortDescription = "Prometheus scrape lists based on PuppetDB"
		parser.WriteManPage(&buf)
		fmt.Printf("%s", buf.String())
		os.Exit(0)
	}
	return
}
