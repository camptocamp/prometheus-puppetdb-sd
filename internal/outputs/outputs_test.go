package outputs

import (
	"github.com/camptocamp/prometheus-puppetdb/internal/config"
	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

var scrapeConfigs = [][]*types.ScrapeConfig{
	[]*types.ScrapeConfig{
		{
			JobName: "node-exporter",
			StaticConfigs: []*types.StaticConfig{
				{
					Targets: []string{
						"server-1.example.com:9100",
					},
					Labels: map[string]string{
						"certname":    "server-1.example.com",
						"environment": "production",
						"team":        "team-1",
					},
				},
				{
					Targets: []string{
						"server-2.example.com:9100",
					},
					Labels: map[string]string{
						"certname":    "server-2.example.com",
						"environment": "development",
						"team":        "team-1",
					},
				},
			},
		},
		{
			JobName: "apache-exporter",
			StaticConfigs: []*types.StaticConfig{
				{
					Targets: []string{
						"server-1.example.com:9117",
					},
					Labels: map[string]string{
						"certname":    "server-1.example.com",
						"environment": "production",
						"team":        "team-2",
					},
				},
			},
		},
	},
	[]*types.ScrapeConfig{
		{
			JobName: "node-exporter",
			StaticConfigs: []*types.StaticConfig{
				{
					Targets: []string{
						"server-1.example.com:9100",
					},
					Labels: map[string]string{
						"certname":    "server-1.example.com",
						"environment": "production",
						"team":        "team-1",
					},
				},
				{
					Targets: []string{
						"server-2.example.com:9100",
					},
					Labels: map[string]string{
						"certname":    "server-2.example.com",
						"environment": "development",
						"team":        "team-2",
					},
				},
			},
		},
	},
}

var expectedOutputs = []map[config.OutputFormat]interface{}{
	map[config.OutputFormat]interface{}{
		config.ScrapeConfigs: `
- job_name: node-exporter
  static_configs:
  - targets:
    - server-1.example.com:9100
    labels:
      certname: server-1.example.com
      environment: production
      team: team-1
  - targets:
    - server-2.example.com:9100
    labels:
      certname: server-2.example.com
      environment: development
      team: team-1
- job_name: apache-exporter
  static_configs:
  - targets:
    - server-1.example.com:9117
    labels:
      certname: server-1.example.com
      environment: production
      team: team-2
`,
		config.StaticConfigs: map[string]string{
			"node-exporter": `
- targets:
  - server-1.example.com:9100
  labels:
    certname: server-1.example.com
    environment: production
    team: team-1
- targets:
  - server-2.example.com:9100
  labels:
    certname: server-2.example.com
    environment: development
    team: team-1
`,
			"apache-exporter": `
- targets:
  - server-1.example.com:9117
  labels:
    certname: server-1.example.com
    environment: production
    team: team-2
`,
		},
		config.MergedStaticConfigs: `
- targets:
  - server-1.example.com:9100
  labels:
    certname: server-1.example.com
    environment: production
    team: team-1
- targets:
  - server-2.example.com:9100
  labels:
    certname: server-2.example.com
    environment: development
    team: team-1
- targets:
  - server-1.example.com:9117
  labels:
    certname: server-1.example.com
    environment: production
    team: team-2
`,
	},
	map[config.OutputFormat]interface{}{
		config.ScrapeConfigs: `
- job_name: node-exporter
  static_configs:
  - targets:
    - server-1.example.com:9100
    labels:
      certname: server-1.example.com
      environment: production
      team: team-1
  - targets:
    - server-2.example.com:9100
    labels:
      certname: server-2.example.com
      environment: development
      team: team-2
`,
		config.StaticConfigs: map[string]string{
			"node-exporter": `
- targets:
  - server-1.example.com:9100
  labels:
    certname: server-1.example.com
    environment: production
    team: team-1
- targets:
  - server-2.example.com:9100
  labels:
    certname: server-2.example.com
    environment: development
    team: team-2
`,
		},
		config.MergedStaticConfigs: `
- targets:
  - server-1.example.com:9100
  labels:
    certname: server-1.example.com
    environment: production
    team: team-1
- targets:
  - server-2.example.com:9100
  labels:
    certname: server-2.example.com
    environment: development
    team: team-2
`,
	},
}
