# Public Go toolchain; build with --platform to select the target architecture.
FROM docker.io/library/golang:1.27.2-bookworm

ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org,direct \
    GOFLAGS=-buildvcs=false

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates curl git jq make openssh-client \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/onsi/ginkgo/ginkgo@v1.16.5 \
    && go install golang.org/x/tools/cmd/goimports@v0.51.0 \
    && go install github.com/mikefarah/yq/v4@v4.52.4 \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0

# Generation can run with the invoking developer's UID against a mounted tree.
RUN mkdir -p /.cache && chmod -R a+rwX /.cache /go
