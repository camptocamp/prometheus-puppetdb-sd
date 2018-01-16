FROM golang:1.9 as builder
COPY . /tmp/
RUN go get -u github.com/jessevdk/go-flags \
              gopkg.in/yaml.v2
RUN make -C /tmp


FROM scratch
COPY --from=builder /tmp/prometheus-puppetdb /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/prometheus-puppetdb"]
VOLUME [ "/etc/prometheus-targets" ]
CMD [""]
