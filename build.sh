#!/usr/bin/env bash
REPO=github.com/bnb-chain/greenfield-storage-provider
Version=`git describe --abbrev=0 --tags --always`
BranchName=`git rev-parse --abbrev-ref HEAD`
CommitID=`git rev-parse HEAD`
BuildTime=`date +%Y-%m-%d\ %H:%M`
CommitDate=`git log -n1 --pretty='format:%cd' --date=format:'%Y%m%d'`

if [ ! -d build  ];then
  mkdir -p build
  mkdir -p build/data
fi

buf generate

go build -ldflags "\
  -extldflags=${EXT_LD_FLAGS}
  -X 'main.Version=${Version}' \
  -X 'main.CommitID=${CommitID}' \
  -X 'main.BranchName=${BranchName}' \
  -X 'main.BuildTime=${BuildTime}' \
  -X '${REPO}/store/bsdb.AppVersion=${Version}' \
  -X '${REPO}/store/bsdb.GitCommit=${CommitID}' \
  -X '${REPO}/store/bsdb.GitCommitDate=${CommitDate}'" \
-o ./build/gnfd-sp cmd/storage_provider/*.go

if [ $? -ne 0 ]; then
    echo "build failed Ooooooh!!!"
    exit 1
else
    echo "build succeed!"
fi
