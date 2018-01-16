DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always)

all: prometheus-puppetdb prometheus-puppetdb.1

prometheus-puppetdb: main.go $(DEPS)
	CGO_ENABLED=0 GOOS=linux \
	  go build -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

prometheus-puppetdb.1: prometheus-puppetdb
	./prometheus-puppetdb -m > $@

clean:
	rm -f prometheus-puppetdb prometheus-puppetdb.1

.PHONY: all clean
