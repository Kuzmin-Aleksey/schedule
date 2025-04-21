set shell := ["powershell.exe", "-c"]

gen-swagger:
     swag init -g cmd/schedule/main.go

gen-api:
    swagger generate server -f docs/swagger.yaml

gen-proto:
    protoc -I proto proto/schedule.proto --go_out=./gen-proto --go_opt=paths=source_relative --go-grpc_out=./gen-proto --go-grpc_opt=paths=source_relative