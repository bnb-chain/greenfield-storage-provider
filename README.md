# Greenfield-storage-provider

Greenfield-Storage-Providers storage service infrastructures provided by either organizations or individuals. They use Greenfield-Storage-Chain as the ledger and the golden data source of meta. Each SP can and will respond to users’ requests to write (upload) and read (download) data, and be the gatekeeper for user rights and authentications.

# Service
## Install-Tools
```shell
make install-tools
```
## Build
```shell
bash build.sh &&
cd build &&

# show version
./gnfd-sp version
Greenfield Storage Provider
    __                                                       _     __
    _____/ /_____  _________ _____ ____     ____  _________ _   __(_)___/ /__  _____
    / ___/ __/ __ \/ ___/ __  / __  / _ \   / __ \/ ___/ __ \ | / / / __  / _ \/ ___/
    (__  ) /_/ /_/ / /  / /_/ / /_/ /  __/  / /_/ / /  / /_/ / |/ / / /_/ /  __/ /
    /____/\__/\____/_/   \__,_/\__, /\___/  / .___/_/   \____/|___/_/\__,_/\___/_/
    /____/       /_/

Version : v0.0.3
Branch  : master
Commit  : e332362ec59724e143725dc5a5a0dacae3be73be
Build   : go1.19.1 darwin amd64 2023-03-13 14:11

# show help
./gnfd-sp help
```

## Deployment
[Deploy SP](docs/tutorial/01-deployment.md)

[Setup Local Test](docs/tutorial/03-localup.md)
