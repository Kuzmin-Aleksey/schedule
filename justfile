set shell := ["pwsh.exe", "-CommandWithArgs"]
set dotenv-load := true

MIGRATION_DIR := "db/migrations"
TEST_DOCKER_COMPOSE := "docker compose -f tests/docker-compose.yml"

# generate api models using swagger
gen-api-model:
    oapi-codegen --package rest --generate models --o pkg/rest/models.gen.go api/openapi.json

gen-test-client:
    oapi-codegen --package rest --generate client --o pkg/rest/apiclient.gen.go api/openapi.json

gen-proto:
    protoc -I proto proto/schedule.proto --go_out=./pkg/grpc --go_opt=paths=source_relative --go-grpc_out=./pkg/grpc --go-grpc_opt=paths=source_relative

lint:
    golangci-lint run -D errcheck

unit-test:
    go test -short -v ./...

goose-create NAME:
    goose -v -dir {{MIGRATION_DIR}} create {{NAME}} sql

goose-up:
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('$MYSQL_SCHEMA')}}?parseTime=true" up
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('$MYSQL_SCHEMA')}}?parseTime=true" status

goose-down:
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('$MYSQL_SCHEMA')}}?parseTime=true" down
    goose -v -dir {{MIGRATION_DIR}} mysql "{{env('MYSQL_USER', 'root')}}:{{env('MYSQL_PASSWORD')}}@/{{env('$MYSQL_SCHEMA')}}?parseTime=true" status


test-infrastructure: test-infrastructure-down
    {{TEST_DOCKER_COMPOSE}} --env-file .env -p tests up --detach --build
    {{TEST_DOCKER_COMPOSE}} logs --follow

test-infrastructure-down:
	{{TEST_DOCKER_COMPOSE}} down --remove-orphans

test:
    go test -v ./...

install_deps:
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
    go install github.com/pressly/goose/v3/cmd/goose@latest