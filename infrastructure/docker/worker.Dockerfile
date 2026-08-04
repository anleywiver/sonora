FROM golang:1.25-alpine AS builder
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
# ffmpeg provides ffprobe, used by the ingest pipeline (Sprint 3) to read
# duration and tags from uploaded audio files. postgresql-client provides
# pg_dump and openssh-client provides scp, both for the Sprint 13
# scheduled backup (ADR 0007).
RUN apk add --no-cache ca-certificates ffmpeg postgresql-client openssh-client
COPY --from=builder /out/worker /usr/local/bin/worker
ENTRYPOINT ["/usr/local/bin/worker"]
