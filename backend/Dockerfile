# syntax=docker/dockerfile:1.7

# Этап сборки
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка под целевую архитектуру
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /build/api ./cmd/api

# Этап запуска
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/api /app/api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/api"]