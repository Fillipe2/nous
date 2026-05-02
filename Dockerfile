FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p bin && go build -o bin/buildnix-backend-webapi ./cmd/webapi

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/buildnix-backend-webapi .
RUN addgroup -S buildnix && adduser -S buildnix -G buildnix && chown -R buildnix:buildnix /app
USER buildnix
EXPOSE 8080
CMD ["./buildnix-backend-webapi"]
