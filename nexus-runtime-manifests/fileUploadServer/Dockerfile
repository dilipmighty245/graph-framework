FROM docker.io/library/golang:1.27.2-bookworm as builder

WORKDIR /app
COPY go.mod go.sum main.go /app/

RUN go build -o server main.go && chmod +x server

FROM gcr.io/nsx-sm/photon:4.0

COPY  --from=builder /app/server /
ENTRYPOINT [ "./server" ]
