GOOS ?= $(shell go env GOOS)

ifeq ($(GOOS),windows)
    IS_EXE := .exe
endif

GOLLERY_BINARY ?= $(GOPATH)/bin/gollery$(IS_EXE)


all: install build run
install:
	@echo "Installing go packages...(this will take a while)"
	dep ensure
	go get -u github.com/go-bindata/go-bindata/...
	go get -u github.com/elazarl/go-bindata-assetfs/...

build-dev:
	@echo "Compiling static assets..."
#	$(GOPATH)/bin/go-bindata-assetfs -pkg gollery -o static.go -debug web/static/...
#	$(GOPATH)/bin/go-bindata-assetfs -pkg gollery -o bindata.go -debug web/...
	@echo "Building..."
	go build -o $(GOLLERY_BINARY) cmd/gollery/main.go
build:
	@echo "Compiling static assets..."
	$(GOPATH)/bin/go-bindata-assetfs -pkg gollery -o static.go web/static/...
	$(GOPATH)/bin/go-bindata -pkg gollery -o template.go web/template/...
	@echo "Building..."
	go build -o $(GOLLERY_BINARY) cmd/gollery/main.go
run:
	@echo "Running the server..."
	./$(GOLLERY_BINARY)
clean:
	$(GOCLEAN)
	rm -f $(GOLLERY_BINARY)
#release: build
#	@echo "Building docker image to cross-compile touchy..."
#	docker build -t touchy .
#	@echo "Removing previous builds..."
#	rm -rf ./touchy_*
#	@echo "Compiling touchy for multiple archs..."
#	docker run -ti -v $$(pwd):/go/src/github.com/odino/touchy touchy
dev: build-dev run