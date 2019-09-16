# Change Log

## [0.11.1](https://github.com/camptocamp/prometheus-puppetdb/tree/0.11.1) (2019-09-16)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.11.0...0.11.1)

**Closed issues:**

- Dockerfile: fix a bug with CMD

## [0.11.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.11.0) (2019-08-21)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.10.0...0.11.0)

**Breaking changes:**
- Project has been renamed to Prometheus PuppetDB SD.
- Kubernetes configmap output has been modified to become a Kubernetes secret output.
- Puppetdb input source has been switched from custom fact to Prometheus::Scrape_job resource.
- Default output format has been changed.

**Implemented enhancements:**

- Fix ineffassign warnings [\#18](https://github.com/camptocamp/prometheus-puppetdb/issues/18)
- Fix golint warnings [\#15](https://github.com/camptocamp/prometheus-puppetdb/issues/15)
- Configuration has been refactored: it is now organized in group of options.
- New output formats have been added: in addition to a unique static_config list, a list of scrape_configs or a list of static_config by job can now be outputted.

**Closed issues:**

- Add Rancher secret output [\#29](https://github.com/camptocamp/prometheus-puppetdb/issues/29)

## [0.10.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.10.0) (2019-05-15)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.9.1...0.10.0)

**Merged pull requests:**

- Add output `external-service` for Kubernetes [\#33](https://github.com/camptocamp/prometheus-puppetdb/pull/33) ([cryptobioz](https://github.com/cryptobioz))
- Move outputs management into a dedicated package [\#32](https://github.com/camptocamp/prometheus-puppetdb/pull/32) ([cryptobioz](https://github.com/cryptobioz))
- Do not stop if a scrape failed [\#31](https://github.com/camptocamp/prometheus-puppetdb/pull/31) ([cryptobioz](https://github.com/cryptobioz))

## [0.9.1](https://github.com/camptocamp/prometheus-puppetdb/tree/0.9.1) (2018-10-30)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.9.0...0.9.1)

**Fixed bugs:**

- Adding an exporter without labels crashes [\#27](https://github.com/camptocamp/prometheus-puppetdb/issues/27)

**Merged pull requests:**

- Check if exporter has labels [\#30](https://github.com/camptocamp/prometheus-puppetdb/pull/30) ([cryptobioz](https://github.com/cryptobioz))

## [0.9.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.9.0) (2018-06-26)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.8.1...0.9.0)

**Merged pull requests:**

- \[WIP\] Add support for custom labels [\#25](https://github.com/camptocamp/prometheus-puppetdb/pull/25) ([cryptobioz](https://github.com/cryptobioz))

## [0.8.1](https://github.com/camptocamp/prometheus-puppetdb/tree/0.8.1) (2018-05-22)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.8.0...0.8.1)

**Implemented enhancements:**

- Remove deprecated backward compatbility feature [\#16](https://github.com/camptocamp/prometheus-puppetdb/issues/16)
- Use vendoring [\#12](https://github.com/camptocamp/prometheus-puppetdb/issues/12)

**Fixed bugs:**

- Fix Travis CI configuration \(or don't use it anymore\) [\#14](https://github.com/camptocamp/prometheus-puppetdb/issues/14)

**Closed issues:**

- Fetch current NameSpace [\#20](https://github.com/camptocamp/prometheus-puppetdb/issues/20)
- How to include the scrape configs in conf.d in Prometheus main configuration? [\#7](https://github.com/camptocamp/prometheus-puppetdb/issues/7)
- Role to target mapping [\#5](https://github.com/camptocamp/prometheus-puppetdb/issues/5)

**Merged pull requests:**

- Refactoring [\#24](https://github.com/camptocamp/prometheus-puppetdb/pull/24) ([cryptobioz](https://github.com/cryptobioz))
- Fix Travis CI \(fix \#14\) [\#23](https://github.com/camptocamp/prometheus-puppetdb/pull/23) ([cryptobioz](https://github.com/cryptobioz))
- Add vendoring \(fix \#12\) [\#22](https://github.com/camptocamp/prometheus-puppetdb/pull/22) ([cryptobioz](https://github.com/cryptobioz))
- Add support to fetch current namespace [\#21](https://github.com/camptocamp/prometheus-puppetdb/pull/21) ([cburki](https://github.com/cburki))
- \[Do not merge yet\] Remove backward compatibility \(Fixes \#16\) [\#17](https://github.com/camptocamp/prometheus-puppetdb/pull/17) ([mcanevet](https://github.com/mcanevet))

## [0.8.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.8.0) (2018-03-23)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.7.1...0.8.0)

## [0.7.1](https://github.com/camptocamp/prometheus-puppetdb/tree/0.7.1) (2018-02-26)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.7.0...0.7.1)

## [0.7.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.7.0) (2018-02-26)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.6.0...0.7.0)

**Merged pull requests:**

- Use relabel\_configs to override scheme and metrics\_path [\#10](https://github.com/camptocamp/prometheus-puppetdb/pull/10) ([mcanevet](https://github.com/mcanevet))

## [0.6.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.6.0) (2018-02-15)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.5.2...0.6.0)

**Closed issues:**

- Connecting to Puppet DB over HTTPS \(TLS\) [\#3](https://github.com/camptocamp/prometheus-puppetdb/issues/3)

**Merged pull requests:**

- V0.6 [\#9](https://github.com/camptocamp/prometheus-puppetdb/pull/9) ([mcanevet](https://github.com/mcanevet))
- Support client TLS auth for secured PuppetDB [\#4](https://github.com/camptocamp/prometheus-puppetdb/pull/4) ([dannyk81](https://github.com/dannyk81))

## [0.5.2](https://github.com/camptocamp/prometheus-puppetdb/tree/0.5.2) (2017-06-13)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.5.1...0.5.2)

## [0.5.1](https://github.com/camptocamp/prometheus-puppetdb/tree/0.5.1) (2017-05-04)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.5.0...0.5.1)

**Merged pull requests:**

- \[WIP\] Allow overrides [\#2](https://github.com/camptocamp/prometheus-puppetdb/pull/2) ([mcanevet](https://github.com/mcanevet))

## [0.5.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.5.0) (2017-04-27)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.4.4...0.5.0)

## [0.4.4](https://github.com/camptocamp/prometheus-puppetdb/tree/0.4.4) (2017-03-22)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.4.3...0.4.4)

## [0.4.3](https://github.com/camptocamp/prometheus-puppetdb/tree/0.4.3) (2017-03-17)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.4.2...0.4.3)

## [0.4.2](https://github.com/camptocamp/prometheus-puppetdb/tree/0.4.2) (2017-03-15)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.4.1...0.4.2)

**Merged pull requests:**

- Exclude EL5 distro \(no prometheus\_writer available\) [\#1](https://github.com/camptocamp/prometheus-puppetdb/pull/1) ([mcanevet](https://github.com/mcanevet))

## [0.4.1](https://github.com/camptocamp/prometheus-puppetdb/tree/0.4.1) (2017-01-31)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.4.0...0.4.1)

## [0.4.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.4.0) (2017-01-12)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.3.0...0.4.0)

## [0.3.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.3.0) (2017-01-12)
[Full Changelog](https://github.com/camptocamp/prometheus-puppetdb/compare/0.2.0...0.3.0)

## [0.2.0](https://github.com/camptocamp/prometheus-puppetdb/tree/0.2.0) (2017-01-12)


\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*
