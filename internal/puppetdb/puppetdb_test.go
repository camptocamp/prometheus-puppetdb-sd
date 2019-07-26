package puppetdb

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/types"
)

var fakeResponse = `
[
	{
		"certname": "server-1.example.com",
		"parameters": {
			"job_name": "node-exporter",
			"targets": [
				"server-1.example.com:9100"
			],
			"labels": {
				"environment": "production",
				"team": "team-1"
			}
		}
	},
	{
		"certname": "server-1.example.com",
		"parameters": {
			"job_name": "apache-exporter",
			"targets": [
				"server-1.example.com:9117"
			],
			"labels": {
				"environment": "production",
				"team": "team-2"
			}
		}
	},
	{
		"certname": "server-2.example.com",
		"parameters": {
			"job_name": "node-exporter",
			"targets": [
				"server-2.example.com:9100"
			],
			"labels": {
				"environment": "development",
				"team": "team-1"
			}
		}
	}
]
`

func testNewClient(t *testing.T, config *config.PuppetDBConfig) {
	client, err := NewClient(config)
	if err != nil {
		assert.FailNow(t, "Failed to create PuppetDB client", err.Error())
	}

	// Test client
	resp, err := client.client.Get(config.URL + "/fake")
	if err != nil {
		assert.FailNow(t, "Failed to issue GET request", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		assert.FailNowf(t, "Unexpected HTTP status code", "Expected %d, but got %d", http.StatusOK, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		assert.FailNow(t, "Failed to read HTTP response body", err.Error())
	}
	assert.Equal(t, []byte(fakeResponse), body)
}

// Try to connect to a PuppetDB with a basic HTTP request
func TestNewClientHTTP(t *testing.T) {
	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/fake" {
				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fakeResponse))
			}
		}),
	)
	defer ts.Close()

	config := config.PuppetDBConfig{
		URL: ts.URL,
	}

	testNewClient(t, &config)
}

// Try to connect to a PuppetDB with a basic HTTPS request
func TestNewClientHTTPS(t *testing.T) {
	// Mock http server
	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/fake" {
				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fakeResponse))
			}
		}),
	)
	defer ts.Close()

	config := config.PuppetDBConfig{
		URL:           ts.URL,
		SSLSkipVerify: true,
	}

	testNewClient(t, &config)
}

// Try to connect to a PuppetDB with an SSL authentication
func TestNewClientSSLAuth(t *testing.T) {
	fakeClientCert := filepath.Join("testdata", "client.pem")
	fakeClientKey := filepath.Join("testdata", "client.key")
	fakeCA := filepath.Join("testdata", "ca.pem")

	// Mock http server
	certpool, err := x509.SystemCertPool()
	if err != nil {
		assert.FailNow(t, "Failed to load system certificates while setting up test server", err.Error())
	}
	if certpool == nil {
		certpool = x509.NewCertPool()
	}

	pem, err := ioutil.ReadFile(fakeCA)
	if err != nil {
		assert.FailNow(t, "Failed to load CA while setting up test server", err.Error())
	}
	if !certpool.AppendCertsFromPEM(pem) {
		assert.FailNow(t, "Failed to append certs from PEM while setting up test server", err.Error())
	}

	ts := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/fake" {
				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fakeResponse))
			}
		}),
	)
	ts.TLS.ClientAuth = tls.RequireAndVerifyClientCert
	ts.TLS.ClientCAs = certpool
	ts.TLS.RootCAs = certpool
	defer ts.Close()

	config := config.PuppetDBConfig{
		URL:           ts.URL,
		KeyFile:       fakeClientKey,
		CertFile:      fakeClientCert,
		CACertFile:    fakeCA,
		SSLSkipVerify: true,
	}

	testNewClient(t, &config)
}

func TestGetResources(t *testing.T) {
	expectedResult := []*types.Resource{
		{
			Certname: "server-1.example.com",
			Parameters: types.Parameters{
				JobName: "node-exporter",
				Targets: []string{
					"server-1.example.com:9100",
				},
				Labels: map[string]string{
					"environment": "production",
					"team":        "team-1",
				},
			},
		},
		{
			Certname: "server-1.example.com",
			Parameters: types.Parameters{
				JobName: "apache-exporter",
				Targets: []string{
					"server-1.example.com:9117",
				},
				Labels: map[string]string{
					"environment": "production",
					"team":        "team-2",
				},
			},
		},
		{
			Certname: "server-2.example.com",
			Parameters: types.Parameters{
				JobName: "node-exporter",
				Targets: []string{
					"server-2.example.com:9100",
				},
				Labels: map[string]string{
					"environment": "development",
					"team":        "team-1",
				},
			},
		},
	}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/pdb/query/v4" {
				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fakeResponse))
			}
		}),
	)
	defer ts.Close()

	config := config.PuppetDBConfig{
		URL: ts.URL,
	}

	client, err := NewClient(&config)
	if err != nil {
		assert.FailNow(t, "Failed to create PuppetDB client", err.Error())
	}

	result, err := client.getResources()
	if err != nil {
		assert.FailNow(t, "Failed to get Puppet resources", err.Error())
	}

	assert.Equal(t, expectedResult, result)
}

// This test is badly written, we should use mocks instead
func TestGetScrapeConfigs(t *testing.T) {
	expectedResult := []*types.ScrapeConfig{
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
	}

	// Mock http server
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/pdb/query/v4" {
				w.Header().Add("Content-Type", "application/json")
				w.Write([]byte(fakeResponse))
			}
		}),
	)
	defer ts.Close()

	cfg := config.PuppetDBConfig{
		URL: ts.URL,
	}

	client, err := NewClient(&cfg)
	if err != nil {
		assert.FailNow(t, "Failed to create PuppetDB client", err.Error())
	}

	result, err := client.GetScrapeConfigs(&config.PrometheusSDConfig{})
	if err != nil {
		assert.FailNow(t, "Failed to get Prometheus scrape configurations", err.Error())
	}

	assert.Equal(t, expectedResult, result)
}
