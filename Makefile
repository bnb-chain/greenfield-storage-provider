SHELL := /bin/bash

.PHONY: all check format vet generate install-tools buf-gen build test tidy clean

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  vet                 to do static check"
	@echo "  build               to create bin directory and build"
	@echo "  generate            to generate code"

format:
	bash script/format.sh
	gofmt -w -l .

proto-format:
	buf format -w

proto-format-check:
	buf format --diff --exit-code

vet:
	go vet ./...

generate:
	go generate ./...

install-tools:
	go install github.com/golang/mock/mockgen@v1.6.0
	go install github.com/bufbuild/buf/cmd/buf@v1.13.1
	go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest

buf-gen:
	rm -rf ./base/types/*/*.pb.go && rm -rf ./modular/metadata/types/*.pb.go && rm -rf ./store/types/*.pb.go
	buf generate

build:
	bash +x ./build.sh

tidy:
	go mod tidy
	go mod verify

# only run unit test, exclude e2e tests
test:
	mockgen -source=core/spdb/spdb.go -destination=core/spdb/spdb_mock.go -package=spdb
	mockgen -source=store/bsdb/database.go -destination=store/bsdb/database_mock.go -package=bsdb
	go test `go list ./... | grep -v /test/`
	# go test -cover ./...

clean:
	rm -rf ./base/types/*/*.pb.go && rm -rf ./modular/metadata/types/*.pb.go && rm -rf ./store/types/*.pb.go

lint:
	golangci-lint run --fix
