.PHONY: build
build:
	go build --o server cmd/server/main.go

.PHONY: test
test:
	go generate ./...
	tern migrate -c migrations/tern.conf -m migrations --database godo_test
	go test --v ./... --count=1  -coverpkg=./internal/... -coverprofile=coverage.out

.PHONY: cover
cover:
	go tool cover -func coverage.out

.PHONY: generage
generate:
	swag fmt
	swag init -d cmd/server/,internal/controller/http/,internal/model/
	go generate ./...
	protoc --go_out=. --go_opt=paths=import \
      --go-grpc_out=. --go-grpc_opt=paths=import \
      ./proto/api/v1/pinger.proto

.PHONY: migrate
migrate:
	tern migrate -c migrations/tern.conf -m migrations

.PHONY: test_migrate
test_migrate:
	tern migrate -c migrations/tern.conf -m migrations --database godo_test

.PHONY: lines
lines:
	git ls-files | xargs wc -l

.PHONY: dock
dock:
	docker-compose
.DEFAULT_GOAL := build
