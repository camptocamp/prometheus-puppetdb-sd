package types

// Exporter contains exporter targets and labels
type Exporter struct {
	URL    string
	Labels map[string]string
}

// Node contains Puppet node informations
type Node struct {
	Certname  string                 `json:"certname"`
	Exporters map[string]interface{} `json:"value"`
}

// StaticConfig contains Prometheus static targets
type StaticConfig struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}
