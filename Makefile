BINARY   := rancher-rackspace-spot-driver
IMAGE    := ghcr.io/teamzuzu/$(BINARY)
TAG      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X main.version=$(TAG)

.PHONY: all build test lint vet image push clean

all: build

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

lint:
	@which golangci-lint > /dev/null || (echo "install golangci-lint first" && exit 1)
	golangci-lint run ./...

vet:
	go vet ./...

image:
	docker build --platform linux/amd64 -t $(IMAGE):$(TAG) .
	docker tag $(IMAGE):$(TAG) $(IMAGE):latest

push: image
	docker push $(IMAGE):$(TAG)
	docker push $(IMAGE):latest

tidy:
	go mod tidy

clean:
	rm -rf bin/
