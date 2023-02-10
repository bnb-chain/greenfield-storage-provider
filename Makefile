SHELL := /bin/bash

.PHONY: all check format vet generate install-tools buf-gen build test tidy clean

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  vet                 to do static check"
	@echo "  build               to create bin directory and build"
	@echo "  generate            to generate code"

format:
	gofmt -w -l .

vet:
	go vet ./...

generate:
	go generate ./...

install-tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.13.1
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

buf-gen:
	buf generate

build:
	bash +x ./build.sh

tidy:
	go mod tidy
	go mod verify

# only run unit test, exclude e2e tests
test:
	go test `go list ./... | grep -v /test/`
	# go test -cover ./...

clean:
	rm -rf ./pkg/types/v1/*.pb.go && rm -rf ./service/types/v1/*.pb.go
