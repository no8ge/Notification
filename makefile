GOPATH := $(shell go env GOPATH)
export GOPATH


BINARY := notify

# Build Rules
.DEFAULT_GOAL := build
build: deps
	go build -v -o $(BINARY)

clean:
	rm -f $(BINARY)

run: build
	@if [ ! -x ./$(BINARY) ]; then echo "Binary not found or not executable"; exit 1; fi
	./$(BINARY)

fmt:
	go fmt ./...

test:
	go test -coverprofile=coverage.out ./tests

coverage:
	@if [ ! -f coverage.out ]; then echo "Coverage profile file not found"; exit 1; fi
	go tool cover -html=coverage.out

deps:
	go mod tidy

docker:
	@sh ./hack/docker.sh || echo "Failed to build docker image"

chart:
	@sh ./hack/chart.sh || echo "Failed to build chart"

.PHONY: build clean run fmt test coverage deps docker chart