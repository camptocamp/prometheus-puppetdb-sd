Prometheus PuppetDB SD
======================

[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/prometheus-puppetdb-sd.svg)](https://hub.docker.com/r/camptocamp/prometheus-puppetdb-sd/)
[![Build Status](https://img.shields.io/travis/camptocamp/prometheus-puppetdb-sd/master.svg)](https://travis-ci.org/camptocamp/prometheus-puppetdb-sd)
[![Coverage Status](https://img.shields.io/coveralls/camptocamp/prometheus-puppetdb-sd.svg)](https://coveralls.io/r/camptocamp/prometheus-puppetdb-sd?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/prometheus-puppetdb-sd)](https://goreportcard.com/report/github.com/camptocamp/prometheus-puppetdb-sd)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


Prometheus PuppetDB SD is a PuppetDB based service discovery tool for Prometheus. It queries PuppetDB to retrieve a list of targets and output Prometheus configuration to scrape the discovered targets.


## Note on built-in Prometheus PuppetDB Service Discovery 

Prometheus introduced [built-in PuppetDB Service Discovery](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#puppetdb_sd_config) in [version 2.31.0](https://github.com/prometheus/prometheus/blob/main/CHANGELOG.md#2310--2021-11-02).

In the future, this project will be deprecated in favor of the built-in option.


## Installing

```shell
$ go get github.com/camptocamp/prometheus-puppetdb-sd
```

## Usage

```shell
Usage:
  prometheus-puppetdb-sd [OPTIONS]

Application Options:
  -V, --version                                                             Display version.
  -m, --manpage                                                             Output manpage.
  -s, --sleep=                                                              Sleep time between queries. (default: 5s) [$SLEEP]

PuppetDB Client Options:
  -u, --puppetdb.url=                                                       PuppetDB base URL. (default: http://puppetdb:8080) [$PUPPETDB_URL]
  -x, --puppetdb.cert-file=                                                 A PEM encoded certificate file. [$PUPPETDB_CERT_FILE]
  -y, --puppetdb.key-file=                                                  A PEM encoded private key file. [$PUPPETDB_KEY_FILE]
  -z, --puppetdb.cacert-file=                                               A PEM encoded CA's certificate file. [$PUPPETDB_CACERT_FILE]
  -k, --puppetdb.ssl-skip-verify                                            Skip SSL verification. [$PUPPETDB_SSL_SKIP_VERIFY]
  -q, --puppetdb.query=                                                     PuppetDB query. (default: resources[certname, parameters] { type = 'Prometheus::Scrape_job' and exported = true }) [$PUPPETDB_QUERY]

Prometheus Service Discovery Options:
      --prometheus.proxy-url=                                               Prometheus target scraping proxy URL. [$PROMETHEUS_PROXY_URL]

Output Configuration:
  -o, --output.method=[stdout|file|k8s-secret]                              Output method. (default: stdout) [$OUTPUT_METHOD]
      --output.format=[scrape-configs|static-configs|merged-static-configs] Output format. (default: scrape-configs) [$OUTPUT_FORMAT]

File Output Configuration:
  -f, --output.file.filename=                                               Output filename. (default: puppetdb-sd.yml) [$OUTPUT_FILENAME]
      --output.file.filename-pattern=                                       Output filename pattern ('*' is the placeholder). (default: *.yml) [$OUTPUT_FILENAME_PATTERN]
      --output.file.directory=                                              Output directory. (default: /etc/prometheus/puppetdb-sd) [$OUTPUT_DIRECTORY]

Kubernetes Secret Output Configuration:
      --output.k8s-secret.secret-name=                                      Kubernetes secret name. [$OUTPUT_K8S_SECRET_NAME]
      --output.k8s-secret.namespace=                                        Kubernetes namespace. [$OUTPUT_K8S_NAMESPACE]
      --output.k8s-secret.object-labels=                                    Labels to add to Kubernetes objects. (default: app.kubernetes.io/name:prometheus-puppetdb-sd) [$OUTPUT_K8S_OBJECT_LABELS]
      --output.k8s-secret.secret-key=                                       Kubernetes secret key. [$OUTPUT_K8S_SECRET_KEY]
      --output.k8s-secret.secret-key-pattern=                               Kubernetes secret key pattern ('*' is the placeholder). [$OUTPUT_K8S_SECRET_KEY_PATTERN]

Help Options:
  -h, --help                                                                Show this help message
```

## How does it work

Prometheus PuppetDB SD works by querying PuppetDB for `Prometheus::Scrape_job` exported resources. These resources comes from the [Prometheus Puppet module](https://github.com/voxpupuli/puppet-prometheus) either by setting the `export_scrape_job` parameter to `true` when using the module's exporter classes or the module's defined type `prometheus::daemon`, or by using the module's defined type `prometheus::scrape_job` directly.

Prometheus PuppetDB SD then build a Prometheus scrape configuration list from the discovered targets and output it using the chosen method and format.
