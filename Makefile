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

build:
	@echo "Building..."
	go-bindata-assetfs -pkg gollery -o assets.go web/...
	go build -o $(GOLLERY_BINARY) cmd/gollery/main.go
build-docker:
	docker build --rm -t gollery:latest .
run:
	@echo "Running the server..."
	./$(GOLLERY_BINARY) start
clean:
	$(GOCLEAN)
	rm -f $(GOLLERY_BINARY)