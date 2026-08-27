FROM golang:1.27-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl \
    && adduser -D -H -u 10001 app

WORKDIR /app
COPY --from=build /out/server /usr/local/bin/server
COPY config.yml /app/config.yml

USER app
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -fsS http://127.0.0.1:8080/health/live || exit 1

ENTRYPOINT ["/usr/local/bin/server"]
