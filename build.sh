#!/usr/bin/env bash
Version=`git describe --abbrev=0 --tags --always`
BranchName=`git rev-parse --abbrev-ref HEAD`
CommitID=`git rev-parse HEAD`
BuildTime=`date +%Y-%m-%d\ %H:%M`

#[[ "-$GOPATH" == "-" ]] && { echo "GOPATH not set"; exit 1; }

if [ ! -d build  ];then
  mkdir -p build
  mkdir -p build/data
fi

make buf-gen

go build -ldflags "\
  -X 'main.Version=${Version}' \
  -X 'main.CommitID=${CommitID}' \
  -X 'main.BranchName=${BranchName}' \
  -X 'main.BuildTime=${BuildTime}'" \
-o ./build/storage_provider cmd/storage_provider/*.go

if [ $? -ne 0 ]; then
    echo "build failed Ooh!!!"
else
    echo "build succeed!"
fi

go build -o ./build/test-storage-provider test/e2e/services/case_driver.go
if [ $? -ne 0 ]; then
    echo "build test-storage-provider failed Ooh!!!"
fi

cp config/config.toml ./build/
