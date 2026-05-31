.PHONY: build clean test test-integration run migrate-up migrate-down migrate-status tidy lint

MIGRATE_IMAGE ?= migrate/migrate
MIGRATE_NETWORK ?= nous_default
MIGRATE_DATABASE_URL ?= postgres://nous:nous@db:5432/nous?sslmode=disable

build:
	@mkdir -p bin
	go build -o bin/nous-webapi ./cmd/webapi

clean:
	rm -rf bin/

test:
	go test ./... -v -race -cover

test-integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo "TEST_DATABASE_URL deve apontar para um Postgres descartavel" && exit 1)
	go test ./internal/infra/database/postgres -run Integration -count=1 -v

run: build
	./bin/nous-webapi

migrate-up:
	docker run --rm --network $(MIGRATE_NETWORK) -v "$(PWD)/migrations/postgres:/migrations" $(MIGRATE_IMAGE) -path=/migrations -database "$(MIGRATE_DATABASE_URL)" up

migrate-down:
	docker run --rm --network $(MIGRATE_NETWORK) -v "$(PWD)/migrations/postgres:/migrations" $(MIGRATE_IMAGE) -path=/migrations -database "$(MIGRATE_DATABASE_URL)" down 1

migrate-status:
	docker run --rm --network $(MIGRATE_NETWORK) -v "$(PWD)/migrations/postgres:/migrations" $(MIGRATE_IMAGE) -path=/migrations -database "$(MIGRATE_DATABASE_URL)" version

tidy:
	go mod tidy

lint:
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	golangci-lint run ./...
