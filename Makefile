DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always --dirty)

all: puppetdb-prometheus

puppetdb-prometheus: main.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@


clean:
	rm -f puppetdb-prometheus

.PHONY: all clean
