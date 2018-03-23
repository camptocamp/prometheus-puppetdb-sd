Prometheus-PuppetDB
===================

[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/prometheus-puppetdb.svg)](https://hub.docker.com/r/camptocamp/prometheus-puppetdb/)
[![Build Status](https://img.shields.io/travis/camptocamp/prometheus-puppetdb/master.svg)](https://travis-ci.org/camptocamp/prometheus-puppetdb)
[![Coverage Status](https://img.shields.io/coveralls/camptocamp/prometheus-puppetdb.svg)](https://coveralls.io/r/camptocamp/prometheus-puppetdb?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/prometheus-puppetdb)](https://goreportcard.com/report/github.com/camptocamp/prometheus-puppetdb)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


Prometheus scrape lists based on PuppetDB.


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
  -o, --output=         Output. One of stdout, file or configmap. (default:
  stdout) [$PROMETHEUS_PUPPETDB_OUTPUT]
  -f, --config-file     Prometheus target file. (default: /etc/prometheus/targets/prometheus-puppetdb/targets.yml) [$PROMETHEUS_PUPPETDB_FILE]
  --configmap           Kubernetes ConfigMap to update. (default: prometheus-puppetdb) [$PROMETHEUS_PUPPETDB_CONFIGMAP]
  --namespace           Kubernetes NameSpace to use. (default: default) [$PROMETHEUS_PUPPETDB_NAMESPACE]
  -s, --sleep=          Sleep time between queries. (default: 5s) [$PROMETHEUS_PUPPETDB_SLEEP]
  -m, --manpage         Output manpage.

Help Options:
  -h, --help            Show this help message
```

## How does it work

Prometheus-puppetdb looks for a fact in PuppetDB ("prometheus_exporters" by default) to generate the list of targets.

The fact must be a hash using exporters as keys and an array of URIs as values,
e.g.:

```
collectd: [http://node.example.com:1234/metrics]
```

You can populate the fact from Puppet using, for example, the `puppetlabs/concat` module:

For example, you can put this in your main profile:

```puppet
concat { '/etc/puppetlabs/facter/facts.d/prometheus_exporters.yaml':
  ensure => present,
}
concat::fragment {'prometheus_exporters':
  target  => '/etc/puppetlabs/facter/facts.d/prometheus_exporters.yaml',
  content => "prometheus_exporters:\n",
  order   => '1',
}
```

Then, in every profile that deploys a Prometheus Exporter:

```puppet
concat::fragment {'prometheus_exporter_collectd':
  target  => '/etc/puppetlabs/facter/facts.d/prometheus_exporters.yaml',
  content => @("END")
  collectd:
    - http://${::fqdn}:9103/metrics
END
  ,
}
```
or
```puppet
concat::fragment {'prometheus_blackbox_exporter':
  target  => '/etc/puppetlabs/facter/facts.d/prometheus_exporters.yaml',
  content => @("END")
  blackbox:
    - http://${::ipaddress}:9115/probe?target=foo.example.com&module=dns_tcp
    - http://${::ipaddress}:9115/probe?target=bar.example.com&module=dns_tcp
END
  ,
}
```
