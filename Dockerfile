FROM golang:1.9 as builder
WORKDIR /go/src/github.com/camptocamp/prometheus-puppetdb
COPY . .
# TODO: use vendoring
RUN go get github.com/jessevdk/go-flags \
           github.com/sirupsen/logrus \
           gopkg.in/yaml.v1
RUN make prometheus-puppetdb

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/prometheus-puppetdb/prometheus-puppetdb /
ENTRYPOINT ["/prometheus-puppetdb"]
CMD [""]
