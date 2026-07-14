# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Resolve dependencies in a separate, cacheable layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary so it runs on a minimal base image without a C toolchain.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---- runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /app/configs ./configs

USER app

EXPOSE 7070

ENTRYPOINT ["./api"]
