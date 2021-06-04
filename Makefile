# Set VERSION to the latest version tag name. Assuming version tags are formatted 'v*'
VERSION := $(shell git describe --always --abbrev=0 --tags --match "v*" $(git rev-list --tags --max-count=1))
BUILD := $(shell git rev-parse $(VERSION))
PROJECTNAME := "termite"
# We pass that to the main module to generate the correct help text
PROGRAMNAME := $(PROJECTNAME)

# Go related variables.
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOBUILD := $(GOBASE)/build
GOFILES := $(shell find . -type f -name '*.go' -not -path './vendor/*')
GOOS_DARWIN := "darwin"
GOOS_LINUX := "linux"
GOOS_WINDOWS := "windows"
GOARCH_AMD64 := "amd64"
GOARCH_ARM64 := "arm64"
GOARCH_ARM := "arm"

MODFLAGS=-mod=readonly

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -X=main.ProgramName=$(PROGRAMNAME)"

# Redirect error output to a file, so we can show it in development mode.
STDERR := $(GOBUILD)/.$(PROJECTNAME)-stderr.txt

# PID file will keep the process id of the server
PID := $(GOBUILD)/.$(PROJECTNAME).pid

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

default: lint format test

ci-checks: lint format test

install: go-get

format: go-format

lint: go-lint

compile:
	@[ -d $(GOBUILD) ] || mkdir -p $(GOBUILD)
	@-touch $(STDERR)
	@-rm $(STDERR)
	@-$(MAKE) -s go-build #2> $(STDERR)
	#@cat $(STDERR) | sed -e '1s/.*/\nError:\n/'  | sed 's/make\[.*/ /' | sed "/^/s/^/     /" 1>&2


test: go-test

clean:
	@-rm $(GOBIN)/$(PROGRAMNAME)* 2> /dev/null
	@-$(MAKE) go-clean

go-lint:
	@echo "  >  Linting source files..."
	go vet $(MODFLAGS) -c=10 `go list $(MODFLAGS) ./...`

go-format:
	@echo "  >  Formating source files..."
	gofmt -s -w $(GOFILES)

go-build: go-get go-build-linux-amd64 go-build-linux-arm64 go-build-darwin-amd64 go-build-windows-amd64 go-build-windows-arm

go-test:
	go test $(MODFLAGS) `go list $(MODFLAGS) ./...`

go-build-linux-amd64:
	@echo "  >  Building linux amd64 binaries..."
	@GOPATH=$(GOPATH) GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_AMD64) GOBIN=$(GOBIN) go build $(MODFLAGS) $(LDFLAGS) -o $(GOBIN)/$(PROGRAMNAME)-$(GOOS_LINUX)-$(GOARCH_AMD64) $(GOBASE)/internal

go-build-linux-arm64:
	@echo "  >  Building linux arm64 binaries..."
	@GOPATH=$(GOPATH) GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_ARM64) GOBIN=$(GOBIN) go build $(MODFLAGS) $(LDFLAGS) -o $(GOBIN)/$(PROGRAMNAME)-$(GOOS_LINUX)-$(GOARCH_ARM64) $(GOBASE)/internal

go-build-darwin-amd64:
	@echo "  >  Building darwin binaries..."
	@GOPATH=$(GOPATH) GOOS=$(GOOS_DARWIN) GOARCH=$(GOARCH_AMD64) GOBIN=$(GOBIN) go build $(MODFLAGS) $(LDFLAGS) -o $(GOBIN)/$(PROGRAMNAME)-$(GOOS_DARWIN)-$(GOARCH_AMD64) $(GOBASE)/internal

go-build-windows-amd64:
	@echo "  >  Building windows amd64 binaries..."
	@GOPATH=$(GOPATH) GOOS=$(GOOS_WINDOWS) GOARCH=$(GOARCH_AMD64) GOBIN=$(GOBIN) go build $(MODFLAGS) $(LDFLAGS) -o $(GOBIN)/$(PROGRAMNAME)-$(GOOS_WINDOWS)-$(GOARCH_AMD64).exe $(GOBASE)/internal

go-build-windows-arm:
	@echo "  >  Building windows arm binaries..."
	@GOPATH=$(GOPATH) GOOS=$(GOOS_WINDOWS) GOARCH=$(GOARCH_ARM) GOBIN=$(GOBIN) go build $(MODFLAGS) $(LDFLAGS) -o $(GOBIN)/$(PROGRAMNAME)-$(GOOS_WINDOWS)-$(GOARCH_ARM).exe $(GOBASE)/internal

go-generate:
	@echo "  >  Generating dependency files..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go mod tidy

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean $(MODFLAGS) $(GOBASE)
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean -modcache



.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo