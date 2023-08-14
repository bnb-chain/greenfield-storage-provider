SHELL := /bin/bash

.PHONY: all build clean format install-tools generate lint test tidy vet buf-gen proto-clean
.PHONY: install-go-test-coverage check-coverage

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  build                 to create build directory and compile sp"
	@echo "  clean                 to remove build directory"
	@echo "  format                to format sp code"
	@echo "  generate              to generate mock code"
	@echo "  install-tools         to install mockgen, buf and protoc-gen-gocosmos tools"
	@echo "  lint                  to run golangci lint"
	@echo "  test                  to run all sp unit tests"
	@echo "  tidy                  to run go mod tidy and verify"
	@echo "  vet                   to do static check"
	@echo "  buf-gen               to use buf to generate pb.go files"
	@echo "  proto-clean           to remove generated pb.go files"
	@echo "  proto-format          to format proto files"
	@echo "  proto-format-check    to check proto files"

build:
	bash +x ./build.sh

check-coverage: install-go-test-coverage
	@go-test-coverage --config=./.testcoverage.yml || true

clean:
	rm -rf ./build

format:
	bash script/format.sh
	gofmt -w -l .

generate:
	go generate ./...

install-go-test-coverage:
	@go install github.com/vladopajic/go-test-coverage/v2@latest

install-tools:
	go install go.uber.org/mock/mockgen@latest
	go install github.com/bufbuild/buf/cmd/buf@v1.25.0
	go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest

lint:
	golangci-lint run --fix

# only run unit test, exclude e2e tests
test:
	mockgen -source=core/spdb/spdb.go -destination=core/spdb/spdb_mock.go -package=spdb
	mockgen -source=store/bsdb/database.go -destination=store/bsdb/database_mock.go -package=bsdb
	go test `go list ./... | grep -v /test/` -coverpkg=./... -covermode=atomic -coverprofile=./coverage.out -timeout 99999s
	# go test -cover ./...

tidy:
	go mod tidy
	go mod verify

vet:
	go vet ./...

buf-gen:
	rm -rf ./base/types/*/*.pb.go && rm -rf ./modular/metadata/types/*.pb.go && rm -rf ./store/types/*.pb.go
	buf generate

proto-clean:
	rm -rf ./base/types/*/*.pb.go && rm -rf ./modular/metadata/types/*.pb.go && rm -rf ./store/types/*.pb.go

proto-format:
	buf format -w

proto-format-check:
	buf format --diff --exit-code
