.PHONY: build
build:
	go build --o server cmd/server/main.go

.PHONY: test
test:
	go test --v ./... --count=1

.PHONY: cover
cover:
	go generate ./...
	go test -coverpkg=./internal/... -coverprofile=coverage.out ./internal/...
	go tool cover -func coverage.out

.PHONY: generage
generate:
	go generate ./...
	protoc --go_out=. --go_opt=paths=import \
      --go-grpc_out=. --go-grpc_opt=paths=import \
      ./proto/api/v1/pinger.proto

.PHONY: migrate
migrate:
	tern migrate -c migrations/tern.conf -m migrations
	tern migrate -c migrations/tern.conf -m migrations --database godo_test

.PHONY: lines
lines:
	git ls-files | xargs wc -l

.DEFAULT_GOAL := build
