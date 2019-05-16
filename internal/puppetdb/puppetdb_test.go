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

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

var fakeResponse = `
[
	{
		"certname": "foo",
		"value": {
			"foo_lorem": [
				{
					"url": "http://127.0.0.1:9103/metrics",
					"labels": {
						"alpha": "beta"
					}
				}
			],
			"foo_ipsum": [
				{
					"url": "http://127.0.0.1:9104/metrics",
					"labels": {
						"gamma": "delta"
					}
				}
			]
		}
	},
	{
		"certname": "bar",
		"value": {
			"bar_lorem": [
				{
					"url": "http://127.0.0.1:9103/metrics",
					"labels": {
						"alpha": "beta"
					}
				}
			],
			"bar_ipsum": [
				{
					"url": "http://127.0.0.1:9104/metrics",
					"labels": {
						"gamma": "delta"
					}
				}
			]
		}
	}
]
`

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

	client, err := NewClient(ts.URL, "", "", "", false)
	assert.Nil(t, err)

	// Test client
	resp, err := client.client.Get(ts.URL + "/fake")
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, []byte(fakeResponse), body)
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

	client, err := NewClient(ts.URL, "", "", "", true)
	assert.Nil(t, err)

	// Test client
	resp, err := client.client.Get(ts.URL + "/fake")
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, []byte(fakeResponse), body)
}

// Try to connect to a PuppetDB with an SSL authentication
func TestNewClientSSLAuth(t *testing.T) {
	fakeClientCert := filepath.Join("testdata", "client.pem")
	fakeClientKey := filepath.Join("testdata", "client.key")
	fakeCA := filepath.Join("testdata", "ca.pem")

	// Mock http server
	certpool, err := x509.SystemCertPool()
	if err != nil {
		assert.FailNow(t, "failed to load system certificates while setting up test server", err.Error())
	}
	if certpool == nil {
		certpool = x509.NewCertPool()
	}

	pem, err := ioutil.ReadFile(fakeCA)
	if err != nil {
		assert.FailNow(t, "failed to load CA while setting up test server", err.Error())
	}
	if !certpool.AppendCertsFromPEM(pem) {
		assert.FailNow(t, "failed to append certs from PEM while setting up test server", err.Error())
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

	client, err := NewClient(ts.URL, fakeClientCert, fakeClientKey, fakeCA, true)
	assert.Nil(t, err)

	// Test client
	resp, err := client.client.Get(ts.URL + "/fake")
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, []byte(fakeResponse), body)
}

// This test is badly written, we should use mocks instead
func TestGetTargets(t *testing.T) {
	expectedResult := []types.StaticConfig{
		types.StaticConfig{
			Targets: []string{
				"127.0.0.1:9103",
			},
			Labels: map[string]string{
				"alpha":        "beta",
				"certname":     "foo",
				"metrics_path": "/metrics",
				"job":          "foo_lorem",
				"scheme":       "http",
			},
		},
	}
	fakeResponse = `
[
	{
		"certname": "foo",
		"value": {
			"foo_lorem": [
				{
					"url": "http://127.0.0.1:9103/metrics",
					"labels": {
						"alpha": "beta"
					}
				}
			]
		}
	}
]
`

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

	client, err := NewClient(ts.URL, "", "", "", false)
	assert.Nil(t, err)

	result, err := client.GetTargets("fake")
	assert.Nil(t, err)

	assert.Equal(t, expectedResult, result)
}
