# syntax=docker/dockerfile:1

FROM docker.io/library/golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w -X github.com/RandomThacker/donna/services/api/internal/buildinfo.Version=${VERSION} -X github.com/RandomThacker/donna/services/api/internal/buildinfo.Commit=${COMMIT} -X github.com/RandomThacker/donna/services/api/internal/buildinfo.BuildTime=${BUILD_TIME}" \
  -o /out/api ./cmd/api && \
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

FROM docker.io/library/alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY migrations /app/migrations
COPY configs /app/configs
COPY scripts/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

ENV MIGRATIONS_PATH=/app/migrations
ENV CONFIG_DIR=/app/configs
EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
