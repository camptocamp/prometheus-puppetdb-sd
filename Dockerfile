FROM golang:1.12 as builder
WORKDIR /go/src/github.com/camptocamp/prometheus-puppetdb-sd
RUN mkdir -p /etc/prometheus/puppetdb-sd/ && chown -R 1001:root /etc/prometheus/puppetdb-sd/ && chmod -R g=u /etc/prometheus/puppetdb-sd/
COPY . .
RUN make prometheus-puppetdb-sd

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/prometheus-puppetdb-sd/prometheus-puppetdb-sd /
COPY --from=builder --chown=1001:0 /etc/prometheus /etc/prometheus
VOLUME /etc/prometheus/puppetdb-sd/
USER 1001
ENTRYPOINT ["/prometheus-puppetdb-sd"]
