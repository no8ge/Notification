GOPATH := $(shell go env GOPATH)
export GOPATH

APPVERSION := $(shell helm show chart chart | awk '/^appVersion:/ {print $$2}')
CHART_NAME := $(shell helm show chart chart | awk '/^name:/ {print $$2}')

.DEFAULT_GOAL := build
build: deps
	go build -v -o $(CHART_NAME)

clean:
	rm -f $(CHART_NAME)
	rm -f $(CHART_NAME)-*.tgz

run: build
	@if [ ! -x ./$(CHART_NAME) ]; then echo "Binary not found or not executable"; exit 1; fi
	./$(CHART_NAME)

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
	helm push $(CHART_NAME)-*.tgz  oci://registry-1.docker.io/no8ge

.PHONY: build clean run fmt test coverage deps docker chart