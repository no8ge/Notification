GOPATH := $(shell go env GOPATH)
export GOPATH

BINARY := notify
APPVERSION := $(shell helm show chart chart | awk '/^appVersion:/ {print $$2}')
CHART_NAME := $(shell helm show chart chart | awk '/^name:/ {print $$2}')

.DEFAULT_GOAL := build
build: deps
	go build -v -o $(BINARY)

clean:
	rm -f $(BINARY)

run: build
	@if [ ! -x ./$(BINARY) ]; then echo "Binary not found or not executable"; exit 1; fi
	./$(BINARY)

test:
	go test -coverprofile=coverage.out ./tests

coverage:
	@if [ ! -f coverage.out ]; then echo "Coverage profile file not found"; exit 1; fi
	go tool cover -html=coverage.out

deps:
	go mod tidy

docker:
	docker buildx build -f Dockerfile --platform linux/amd64 -t no8ge/$(CHART_NAME):$(APPVERSION) . --push

chart:
	helm package chart 
	helm push notify-*.tgz  oci://registry-1.docker.io/no8ge

.PHONY: build clean run fmt test coverage deps docker chart