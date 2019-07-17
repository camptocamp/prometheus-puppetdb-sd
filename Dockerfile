FROM golang:1.11 as builder
WORKDIR /go/src/github.com/camptocamp/prometheus-puppetdb
COPY . .
RUN make prometheus-puppetdb
RUN mkdir /data && touch /data/targets.yml && chown -R 1001.root /data && chmod -R g=u /data

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/prometheus-puppetdb/prometheus-puppetdb /
COPY --from=builder /data/targets.yml /data/targets.yml
VOLUME /data
USER 1001
ENTRYPOINT ["/prometheus-puppetdb"]
CMD [""]
