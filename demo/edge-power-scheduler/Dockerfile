FROM --platform=linux/amd64 golang:1.27.3-bookworm as builder
WORKDIR /app
COPY . .
RUN go mod download && make build 

FROM --platform=linux/amd64 golang:1.27.3-bookworm
WORKDIR /app
COPY --from=0 /app/bin/amd64/* /bin/
