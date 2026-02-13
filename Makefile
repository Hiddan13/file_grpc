.PHONY: proto test build clean run-server deps fmt

PROTO_DIR = api/proto
PROTO_OUT = api/proto/pb

deps:
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto:
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/file.proto

build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/client cmd/client/main.go

run-server:
	go run cmd/server/main.go

test:
	go test -v -race ./...

clean:
	rm -rf bin $(PROTO_OUT)
	go clean -cache

fmt:
	go fmt ./...