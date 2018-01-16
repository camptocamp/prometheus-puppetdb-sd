FROM golang:1.9 as builder
COPY main.go Makefile /tmp/
RUN go get -u github.com/jessevdk/go-flags \
              gopkg.in/yaml.v2
RUN make -C /tmp


FROM scratch
COPY --from=builder /tmp/prometheus-puppetdb /
ENTRYPOINT ["/prometheus-puppetdb"]
VOLUME [ "/etc/prometheus-targets" ]
CMD [""]
