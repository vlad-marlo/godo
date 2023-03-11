.PHONY: build
build:
	go build --o server cmd/server/main.go

.PHONY: t
t:
	go generate ./...
	tern migrate --config migrations/tern.conf --migrations migrations --database godo_test
	go test --v ./... --count=1 --coverpkg=./internal/... --coverprofile=coverage.out --test.short=true

.PHONY: test
test:
	go generate ./...
	tern migrate --config migrations/tern.conf --migrations migrations --database godo_test
	go test --v ./... --count=1 --coverpkg=./internal/... --coverprofile=coverage.out

.PHONY: c
c:
	go tool cover --func coverage.out

.PHONY: gen
gen:
	swag fmt
	swag init --d cmd/server/,internal/controller/http/,internal/model/
	go generate ./...
	protoc --go_out=. --go_opt=paths=import \
      --go-grpc_out=. --go-grpc_opt=paths=import \
      ./proto/api/v1/pinger.proto

.PHONY: migrate
migrate:
	tern migrate --c migrations/tern.conf --m migrations

.PHONY: tm
tm:
	tern migrate --c migrations/tern.conf --m migrations --database godo_test

.PHONY: lines
lines:
	git ls-files | xargs wc -l

.PHONY: dock
dock:
	docker build . --file=infra/httpserver.dockerfile --tag="marlooooo/godo_backend:http-latest"
	docker build . --file=infra/grpcserver.dockerfile --tag="marlooooo/godo_backend:grpc-latest"
	docker build . --file=infra/server.dockerfile --tag="marlooooo/godo_backend:latest"

.PHONY: run
run:
	docker build . --file=infra/httpserver.dockerfile --tag="marlooooo/godo_backend:http-latest"
	docker compose --file=infra/docker-compose.yml up --d

.PHONY: push
push:
	docker push "marlooooo/godo_backend" --a
.DEFAULT_GOAL := build
