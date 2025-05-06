set shell := ["powershell.exe", "-c"]

# generate api models using swagger
gen-api-model:
     oapi-codegen --package models --generate models --o internal/controller/httpHandler/models/models.go api/openapi.json

gen-proto:
    protoc -I proto proto/schedule.proto --go_out=./internal/server/grpcServer/gen --go_opt=paths=source_relative --go-grpc_out=./internal/server/grpcServer/gen --go-grpc_opt=paths=source_relative

lint:
    golangci-lint run -D errcheck

unit-test:
    go test -v ./...

install_deps:
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest