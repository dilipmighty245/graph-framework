FROM docker.io/library/golang:1.27.3-bookworm

ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct \
    GOFLAGS=-buildvcs=false

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates git jq make openssh-client \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/onsi/ginkgo/ginkgo@v1.16.5 \
    && go install golang.org/x/tools/cmd/goimports@v0.51.0 \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
