############################################################
# Dockerfile to build golang Installed Containers

# Based on alpine

############################################################

FROM golang:1.21 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY="https://goproxy.cn,direct" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM alpine:3.13

RUN mkdir /notify
COPY --from=builder /src/notify /notify

EXPOSE 8081
WORKDIR /notify
CMD ["/notify"]
