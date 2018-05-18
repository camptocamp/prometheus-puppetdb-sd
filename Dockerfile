FROM golang:1.9 as builder
WORKDIR /go/src/github.com/camptocamp/prometheus-puppetdb
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure -vendor-only
RUN make prometheus-puppetdb

FROM scratch
COPY --from=builder /go/src/github.com/camptocamp/prometheus-puppetdb/prometheus-puppetdb /
ENTRYPOINT ["/prometheus-puppetdb"]
CMD [""]
