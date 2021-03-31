gen:
	protoc --go_opt=paths=source_relative --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_opt=paths=source_relative --go-grpc_out=. proto/*.proto

build:
	go build -o service .
