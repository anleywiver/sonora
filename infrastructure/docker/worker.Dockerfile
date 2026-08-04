FROM golang:1.24-alpine AS builder
WORKDIR /src

# GOWORK=off: go.work lists all workspace members (including apps/backend),
# which isn't copied into this build context. Each app resolves go-core via
# its own go.mod replace directive instead.
ENV GOWORK=off
COPY libs/go-core libs/go-core
COPY apps/worker apps/worker

WORKDIR /src/apps/worker
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/worker /usr/local/bin/worker
ENTRYPOINT ["/usr/local/bin/worker"]
