Prometheus-PuppetDB
===================

[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/prometheus-puppetdb.svg)](https://hub.docker.com/r/camptocamp/prometheus-puppetdb/)
[![Build Status](https://img.shields.io/travis/camptocamp/prometheus-puppetdb/master.svg)](https://travis-ci.org/camptocamp/prometheus-puppetdb)
[![Coverage Status](https://img.shields.io/coveralls/camptocamp/prometheus-puppetdb.svg)](https://coveralls.io/r/camptocamp/prometheus-puppetdb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/prometheus-puppetdb)](https://goreportcard.com/report/github.com/camptocamp/prometheus-puppetdb)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


Prometheus scape lists based on PuppetDB.


## Installing

```shell
$ go get github.com/camptocamp/prometheus-puppetdb
```

## Usage

```shell
Usage:
  prometheus-puppetdb [OPTIONS]

Application Options:
  -V, --version         Display version.
  -u, --puppetdb-url=   PuppetDB base URL. (default: http://puppetdb:8080) [$PROMETHEUS_PUPPETDB_URL]
  -x, --cert-file=      A PEM encoded certificate file. (default: certs/client.pem) [$PROMETHEUS_CERT_FILE]
  -y, --key-file=       A PEM encoded private key file. (default: certs/client.key) [$PROMETHEUS_KEY_FILE]
  -z, --cacert-file=    A PEM encoded CA's certificate file. (default: certs/cacert.pem) [$PROMETHEUS_CACERT_FILE]
  -k, --ssl-skip-verify Skip SSL verification.
  -q, --puppetdb-query= PuppetDB query. (default: facts[certname, value] { name='prometheus_exporters' and nodes { deactivated is null } }) [$PROMETHEUS_PUPPETDB_QUERY]
  -c, --config-dir=     Prometheus config dir. (default: /etc/prometheus) [$PROMETHEUS_CONFIG_DIR]
  -s, --sleep=          Sleep time between queries. (default: 5s) [$PROMETHEUS_PUPPETDB_SLEEP]
  -m, --manpage         Output manpage.

Help Options:
  -h, --help            Show this help message
```
