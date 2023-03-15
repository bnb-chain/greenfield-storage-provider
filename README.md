# Greenfield Storage Provider

Storage Provider (abbreviated SP) is storage service infrastructure providers. It uses Greenfield as the ledger 
and the single source of truth. Each SP can and will respond to users' requests to write (upload) and read (download) 
data, and serve as the gatekeeper for user rights and authentications.

## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## SP Core(coming soon)


## Quick Started

*Note*: Requires [Go 1.18+](https://go.dev/dl/)

### Install-Tools
```shell
make install-tools
```
### Build
```shell
make build &&
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
### Configuration

#### Make configuration template
```shell
# dump default configuration
./gnfd-sp config.dump
```

#### Edit configuration 
```shell
# start service list
Service = ["gateway", "uploader", "downloader", "challenge", "stonenode", "syncer", "signer", "metadata", "manager"]
# sp operator address 
SpOperatorAddress = ""
Domain = "gnfd.nodereal.com"
# gate listen http address
[HTTPAddress]
gateway = "localhost:9033"
# service listen address
[GRPCAddress]
challenge = "localhost:9333"
downloader = "localhost:9233"
metadata = "localhost:9733"
signer = "localhost:9633"
stonenode = "localhost:9433"
syncer = "localhost:9533"
uploader = "localhost:9133"
# SQL configuration
[SpDBConfig]
User = "root"
Passwd = "test_pwd"
Address = "localhost:3306"
Database = "storage_provider_db"
# piece store configuration
[PieceStoreConfig]
Shards = 0
[PieceStoreConfig.Store]
# default use local file system 
Storage = "file"
BucketURL = "./data"
# greenfiel chain configuration
[ChainConfig]
ChainID = "greenfield_9000-1741"
[[ChainConfig.NodeAddr]]
GreenfieldAddresses = ["localhost:9090"]
TendermintAddresses = ["http://localhost:26750"]
```

#### Start SP

```shell
# start sp
./gnfd-sp --config ${config_file_path}
```

#### Add SP to Greenfield
[Add SP to Greenfield](https://github.com/bnb-chain/greenfield/blob/master/docs/cli/storage-provider.md)

## Document
* [Greenfield Whitepaper](https://github.com/bnb-chain/greenfield-whitepaper): the official Greenfield Whitepaper.
* [SP Introduce](docs/readme.md): the Greenfield Storage Provider documents.
* [SP Deployment](docs/tutorial/01-deployment.md): the detailed introduction to deploying sp.
* [SP Local Setup](docs/run-book/03-local.toml): the introduction to set up local SP env for testing.

## Related Projects
* [Greenfield](https://github.com/bnb-chain/greenfield): the Golang implementation of the Greenfield Blockchain.
* [Greenfield-Common](https://github.com/bnb-chain/greenfield-common): the Greenfield common package.
* [reedsolomon](https://github.com/klauspost/reedsolomon): the Reed-Solomon Erasure package in prue Go, with speeds exceeding 1GB/s/cpu core.
* [Greenfield Go SDK](https://github.com/bnb-chain/greenfield-go-sdk): the Greenfield SDK, interact with SP, Greenfield and Tendermint.


## Contribution
Thank you for considering to help out with the source code! We welcome contributions from 
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to Greenfield Storage Provider, please fork, fix, commit and 
send a pull request for the maintainers to review and merge into the main code base. 
If you wish to submit more complex changes though, please check up with the core devs first 
through github issue(going to have a discord channel soon) to ensure those changes are in 
line with the general philosophy of the project and/or get some early feedback which can make 
both your efforts much lighter as well as our review and merge procedures quick and simple.
