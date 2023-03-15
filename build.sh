#!/usr/bin/env bash
Version=`git describe --abbrev=0 --tags --always`
BranchName=`git rev-parse --abbrev-ref HEAD`
CommitID=`git rev-parse HEAD`
BuildTime=`date +%Y-%m-%d\ %H:%M`

if [ ! -d build  ];then
  mkdir -p build
  mkdir -p build/data
fi

buf generate

go build -ldflags "\
  -X 'main.Version=${Version}' \
  -X 'main.CommitID=${CommitID}' \
  -X 'main.BranchName=${BranchName}' \
  -X 'main.BuildTime=${BuildTime}'" \
-o ./build/gnfd-sp cmd/storage_provider/*.go

if [ $? -ne 0 ]; then
    echo "build failed Ooooooh!!!"
else
    echo "build succeed!"
fi
