FROM golang:1.26-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p bin && go build -o bin/buildnix-backend-webapi ./cmd/webapi

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/buildnix-backend-webapi .
RUN addgroup -S buildnix && adduser -S buildnix -G buildnix && chown -R buildnix:buildnix /app
USER buildnix
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8080/health || exit 1
CMD ["./buildnix-backend-webapi"]
