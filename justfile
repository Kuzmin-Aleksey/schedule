set shell := ["powershell.exe", "-c"]

# generate swagger
gen-swagger:
     swag init -g cmd/schedule/main.go

# generate api models using swagger
gen-api-model:
    swagger generate model -f docs/swagger.json -t internal/controller/httpHandler

gen-proto:
    protoc -I proto proto/schedule.proto --go_out=./gen-proto --go_opt=paths=source_relative --go-grpc_out=./gen-proto --go-grpc_opt=paths=source_relative