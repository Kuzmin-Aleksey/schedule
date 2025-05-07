set shell := ["powershell.exe", "-c"]

MIGRATION_DIR := "db/migrations"

# generate api models using swagger
gen-api-model:
     oapi-codegen --package models --generate models --o internal/server/rest/models/models.go api/openapi.json

gen-proto:
    protoc -I proto proto/schedule.proto --go_out=./internal/server/grpcServer/gen --go_opt=paths=source_relative --go-grpc_out=./internal/server/grpcServer/gen --go-grpc_opt=paths=source_relative

lint:
    golangci-lint run -D errcheck

unit-test:
    go test -v ./...

goose-create NAME:
    goose -v -dir {{MIGRATION_DIR}} create {{NAME}} sql

goose-up USER PASSWORD DB:
    goose -v -dir {{MIGRATION_DIR}} mysql "{{USER}}:{{PASSWORD}}@/{{DB}}?parseTime=true" up
    goose -v -dir {{MIGRATION_DIR}} mysql "{{USER}}:{{PASSWORD}}@/{{DB}}?parseTime=true" status

install_deps:
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest