VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOLDFLAGS  = -s -w
GOLDFLAGS += -X main.buildVersion=$(VERSION)
GOLDFLAGS += -X main.buildTime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

runningman: runningman.go
	CGO_ENABLED=0 go build -o runningman $(GOFLAGS) *.go
