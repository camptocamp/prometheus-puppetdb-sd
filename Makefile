DEPS = $(wildcard */*/*/*.go)
VERSION = $(shell git describe --always --tags)

all: lint test prometheus-puppetdb-sd prometheus-puppetdb-sd.1

prometheus-puppetdb-sd: main.go $(DEPS)
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux \
	  go build -mod=vendor -a \
		  -ldflags="-X main.version=$(VERSION)" \
	    -installsuffix cgo -o $@ $<
	strip $@

prometheus-puppetdb-sd.1: prometheus-puppetdb-sd
	./prometheus-puppetdb-sd -m > $@

clean:
	rm -f prometheus-puppetdb-sd prometheus-puppetdb-sd.1

lint:
	go vet $<
	@GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck
	staticcheck -tests ./...
	@GO111MODULE=off go get -u golang.org/x/lint/golint
	@for file in $$(go list ./... | grep -v '_workspace/' | grep -v 'vendor'); do \
		export output="$$(golint $${file} | grep -v 'type name will be used as docker.DockerInfo')"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
	done; \
	exit $${status:-0}

vendor:
	go mod vendor

test:
	go test -cover -coverprofile=coverage -v ./...

.PHONY: all lint clean test
