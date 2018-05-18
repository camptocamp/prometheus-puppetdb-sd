DEPS = $(wildcard */*.go)
VERSION = $(shell git describe --always --dirty)

all: imports lint vet prometheus-puppetdb prometheus-puppetdb.1

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

lint:
	@ go get -v github.com/golang/lint/golint
	@for file in $$(git ls-files '*.go' | grep -v '_workspace/'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vet: main.go
	go vet $<

imports: main.go
	dep ensure -vendor-only
	goimports -d $<

.PHONY: all lint clean
