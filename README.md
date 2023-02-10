# Greenfield-storage-provider

Greenfield-Storage-Providers storage service infrastructures provided by either organizations or individuals. They use GreenField-Storage-Chain as the ledger and the golden data source of meta. Each SP can and will respond to users’ requests to write (upload) and read (download) data, and be the gatekeeper for user rights and authentications.

# Service
## Install-Tools
```shell
make install-tools
```
## Build
```shell
bash build.sh
```
## Quick Deploy
```shell
cd build
# print version
./gnfd-sp --version
# setup secondary sps in the test-env directory(syncer), notice: only run once at first
./setup-test-env
# run primary sp(gateway/uploader/downloader/stonehub/stonenode/syncer)
./gnfd-sp -config ./config.toml
# run cases, request to primary sp
./test-gnfd-sp
```
