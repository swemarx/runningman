VERSION   := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GO_LDFLAGS  = -s -w
GO_LDFLAGS += -X main.buildVersion=$(VERSION)
GO_LDFLAGS += -X main.buildTime=$(BUILDTIME)
GO_FLAGS    = -ldflags "$(GO_LDFLAGS)"

.PHONY: all
all: runningman

runningman: *.go
	CGO_ENABLED=0 go build -o runningman $(GO_FLAGS) *.go

.PHONY: clean
clean:
	rm -f ./runningman
