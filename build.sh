#!/usr/bin/env bash
Version=`git describe --abbrev=0 --tags`
BranchName=`git rev-parse --abbrev-ref HEAD`
CommitID=`git rev-parse HEAD`
BuildTime=`date +%Y-%m-%d\ %H:%M`

#[[ "-$GOPATH" == "-" ]] && { echo "GOPATH not set"; exit 1; }

buf generate

go build -ldflags "\
-X 'main.Version=${Version}' \
-X 'main.CommitID=${CommitID}' \
-X 'main.BranchName=${BranchName}' \
-X 'main.BuildTime=${BuildTime}'" \
-o storage_provider cmd/storage_provider/*.go

if [ ! -d build  ];then
  mkdir build
fi
mv storage_provider ./build
