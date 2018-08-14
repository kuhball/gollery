# inspired by github.com/jessfraz - please don't kill me
# TODO: this is a fucking mess - clean this up

GOOS ?= $(shell go env GOOS)
SHOW := cat

ifeq ($(GOOS),windows)
    IS_EXE := .exe
    SHOW = type
endif

VERSION := $(shell $(SHOW) VERSION.txt)

GOLLERY_BINARY ?= $(GOPATH)/bin/gollery$(IS_EXE)

# Set an output prefix, which is the local directory if not specified
PREFIX:=${CURDIR}

# Setup name variables for the package/tool
NAME := gollery
PKG := github.com/scouball/$(NAME)

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/cross

GOOSARCHES = darwin/amd64 darwin/386 freebsd/amd64 freebsd/386 linux/arm linux/arm64 linux/amd64 linux/386 windows/amd64 windows/386

.PHONY: all
 all: install build

.PHONY: install
install:
	@echo "Installing go packages...(this will take a while)"
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/go-bindata/go-bindata/...
	go get -u github.com/elazarl/go-bindata-assetfs/...
	dep ensure

.PHONY: build
build:
	@echo "Building..."
	go-bindata-assetfs -pkg gollery -o assets.go -ignore=robots.txt web/...
	go build -o $(GOLLERY_BINARY) cmd/gollery/main.go

.PHONY: build-robots
build-robots:
	@echo "Building with robots.txt..."
	go-bindata-assetfs -pkg gollery -o assets.go web/...
	go build -o $(GOLLERY_BINARY) cmd/gollery/main.go

.PHONY: build-docker
build-docker:
	docker build --rm -t gollery:latest .

define buildpretty
mkdir -p $(BUILDDIR)/$(1)/$(2);
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build \
	 -o $(BUILDDIR)/$(1)/$(2)/$(NAME) \
	 -a .;
md5sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).md5;
sha256sum $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).sha256;
endef

.PHONY: cross
cross: cmd/gollery/main.go VERSION.txt ## Builds the cross-compiled binaries, creating a clean directory structure (eg. GOOS/GOARCH/binary)
	@echo "+ $@"
	go-bindata-assetfs -pkg gollery -o assets.go web/...;
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildpretty,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

define buildrelease
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a .;
md5sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).md5;
sha256sum $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
endef

.PHONY: release
release: cmd/gollery/main.go VERSION.txt ## Builds the cross-compiled binaries, naming them in such a way for release (eg. binary-GOOS-GOARCH)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildrelease,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))


.PHONY: run
run:
	@echo "Running the server..."
	$(GOLLERY_BINARY) start

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(GOLLERY_BINARY)