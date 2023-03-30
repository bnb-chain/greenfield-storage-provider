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
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/bufbuild/buf/cmd/buf@v1.13.1
	go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest


buf-gen:
	buf generate

build:
	bash +x ./build.sh

tidy:
	go mod tidy
	go mod verify

# only run unit test, exclude e2e tests
test:
	mockgen -source=store/sqldb/database.go -destination=store/sqldb/database_mock.go -package=sqldb
	go test `go list ./... | grep -v /test/`
	# go test -cover ./...

clean:
	rm -rf ./service/types/*.pb.go && rm -rf ./service/*/types/*.pb.go
