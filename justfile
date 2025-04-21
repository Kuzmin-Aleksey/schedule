set shell := ["powershell.exe", "-c"]

gen-swagger:
     swag init -g cmd/schedule/main.go

gen-proto:
    protoc -I proto proto/schedule.proto --go_out=./gen-proto --go_opt=paths=source_relative --go-grpc_out=./gen-proto --go-grpc_opt=paths=source_relative