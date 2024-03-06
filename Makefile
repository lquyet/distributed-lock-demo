PROTOC_LINUX_VERSION = 3.11.4
PROTOC_LINUX_ZIP = protoc-$(PROTOC_LINUX_VERSION)-linux-x86_64.zip

.PHONY: install-go-tools install-protoc gen-proto migrate-up migrate-down-1 run lint test get-next-err-code check-sql

install-protoc:
	curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_LINUX_VERSION)/$(PROTOC_LINUX_ZIP)
	sudo unzip -o $(PROTOC_LINUX_ZIP) -d /usr/local bin/protoc
	sudo unzip -o $(PROTOC_LINUX_ZIP) -d /usr/local 'include/*'
	rm -f $(PROTOC_LINUX_ZIP)
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

test:
	go test -p 1 -v ./...

run:
	go run server/cmd/server/main.go start
