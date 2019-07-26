package types

// Resource represents a Puppet resource
type Resource struct {
	Certname   string     `json:"certname"`
	Parameters Parameters `json:"parameters"`
}

// Parameters represents the paramaters of a Puppet resource
type Parameters struct {
	JobName string            `json:"job_name"`
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

// ScrapeConfig represents a Prometheus scrape_config
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config
type ScrapeConfig struct {
	JobName       string          `yaml:"job_name"`
	ProxyURL      string          `yaml:"proxy_url,omitempty"`
	StaticConfigs []*StaticConfig `yaml:"static_configs"`
}

// StaticConfig represents a Prometheus static_config
// See https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config
type StaticConfig struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}
