FROM golang:1.12 as builder
WORKDIR /go/src/github.com/camptocamp/prometheus-puppetdb-sd
COPY . .
RUN make prometheus-puppetdb-sd
RUN mkdir /data && touch /data/targets.yml && chown -R 1001.root /data && chmod -R g=u /data

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/prometheus-puppetdb-sd/prometheus-puppetdb-sd /
COPY --from=builder /data/targets.yml /data/targets.yml
VOLUME /data
USER 1001
ENTRYPOINT ["/prometheus-puppetdb-sd"]
CMD [""]
