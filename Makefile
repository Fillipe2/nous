.PHONY: build clean test run migrate-up migrate-down migrate-status tidy lint

build:
	@mkdir -p bin
	go build -o bin/buildnix-backend-webapi ./cmd/webapi

clean:
	rm -rf bin/

test:
	go test ./... -v -race -cover

run: build
	./bin/buildnix-backend-webapi

migrate-up:
	@which goose > /dev/null || go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations/postgres postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations/postgres postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations/postgres postgres "$(DATABASE_URL)" status

tidy:
	go mod tidy

lint:
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	golangci-lint run ./...
