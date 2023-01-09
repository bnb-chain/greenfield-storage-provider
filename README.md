# Inscription-storage-provider

Inscription-storage-provider is storage service infrastructures provided by either organizations or individuals. They use Inscription as the ledger and the golden data source of meta. Each SP can and will respond to users’ requests to write (upload) and read (download) data, and be the gatekeeper for user rights and authentications. 

# Service
## Build
```shell
bash build.sh
```
## Deploy
```shell
# Print Version
./build/storage_provider -v
# Run Services
./build/storage_provider -config config/config.toml
```
