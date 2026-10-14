ARG BUILDER_TAG

FROM $BUILDER_TAG
ADD compiler.tar /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler
ADD cmd/nexus-openapi-gen /go/bin/nexus-openapi-gen
ADD cmd/gqlgen /go/bin/gqlgen

WORKDIR /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler
RUN mv .git .gitmodules kube-openapi gqlgen nexus install-validator common-library  ..
RUN git config --global --add safe.directory '*' && git config --global --add url."https://github.com/".insteadOf "git@github.com:"
RUN make init_submodules

ENV GOCACHE=/root/.cache/go-build

WORKDIR /go/src/
RUN --mount=type=cache,target="/root/.cache/go-build" mkdir nexustmp && cd nexustmp && go mod init && \
    /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/scripts/pin_deps.sh && \
    /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/scripts/pin_graphql_build_version.sh && \
    go mod download || go mod download

WORKDIR /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler/_deps/github.com/kubernetes/code-generator
RUN --mount=type=cache,target="/root/.cache/go-build"  go mod download  || go mod download

WORKDIR /go/src/github.com/dilipmighty245/graph-framework-for-microservices/install-validator
RUN --mount=type=cache,target="/root/.cache/go-build"  CGO_ENABLED=0 go build --trimpath -o install.so ./cmd/validate

WORKDIR /go/src/github.com/dilipmighty245/graph-framework-for-microservices/compiler
RUN --mount=type=cache,target="/root/.cache/go-build" go mod download  || go mod download

CMD make preparser
