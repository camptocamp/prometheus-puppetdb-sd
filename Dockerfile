FROM golang:1.11 as builder
WORKDIR /go/src/github.com/camptocamp/prometheus-puppetdb
COPY . .
RUN make prometheus-puppetdb

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/prometheus-puppetdb/prometheus-puppetdb /
ENTRYPOINT ["/prometheus-puppetdb"]
CMD [""]
