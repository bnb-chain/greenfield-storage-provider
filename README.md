# Greenfield Storage Provider

Greenfield Storage Provider (abbreviated SP) is storage service infrastructure provider. It uses Greenfield as the ledger 
and the single source of truth. Each SP can and will respond to users' requests to write (upload) and read (download) 
data, and serve as the gatekeeper for user rights and authentications.

## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## SP Core
SPs store the objects' real data, i.e. the payload data. Each SP runs its own object storage system. Similar to Amazon 
S3 and other object store systems, the objects stored on SPs are immutable. The users may delete and re-create the object 
(under the different ID, or under the same ID after certain publicly declared settings), but they cannot modify it.

SPs have to register themselves first by depositing on the Greenfield blockchain as their "Service Stake". Greenfield 
validators will go through a dedicated governance procedure to vote for the SPs of their election. SPs are encouraged 
to advertise their information and prove to the community their capability, as SPs have to provide a professional storage 
system with high-quality SLA.

SPs provide publicly accessible APIs for users to upload, download, and manage data. These APIs are very similar to Amazon 
S3 APIs so that existing developers may feel familiar enough to write code for it. Meanwhile, they provide each other REST 
APIs and form another white-listed P2P network to communicate with each other to ensure data availability and redundancy. 
There will also be a P2P-based upload/download network across SPs and user-end client software to facilitate easy connections 
and fast data download, which is similar to BitTorrent.

Among the multiple SPs that one object is stored on, one SP will be the "Primary SP", while the others are "Secondary SP".

When users want to write an object into Greenfield, they or the client software they use must specify the primary SP. Primary 
SP should be used as the only SP to download the data. Users can change the primary SP for their objects later if they are not 
satisfied with its service.


## Quick Started

*Note*: Requires [Go 1.18+](https://go.dev/dl/)

### Install-Tools
```shell
make install-tools
```
### Build
```shell
# build gnfd-sp
make builc && cd build 

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
```toml
# start service list
Service = ["gateway", "uploader", "downloader", "challenge", "tasknode", "receiver", "signer", "blocksyncer", "metadata", "manager"]
# sp operator address 
SpOperatorAddress = ""
# service endpoint for other to connect
[Endpoint]
challenge = "localhost:9333"
downloader = "localhost:9233"
gateway = "gnfd.nodereal.com"
metadata = "localhost:9733"
receiver = "localhost:9533"
signer = "localhost:9633"
tasknode = "localhost:9433"
uploader = "localhost:9133"
# service listen address
[ListenAddress]
challenge = "localhost:9333"
downloader = "localhost:9233"
gateway = "localhost:9033"
metadata = "localhost:9733"
receiver = "localhost:9533"
signer = "localhost:9633"
tasknode = "localhost:9433"
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
# signer configuration
[SignerCfg]
GRPCAddress = "localhost:9633"
APIKey = ""
WhitelistCIDR = ["127.0.0.1/32"]
GasLimit = 210000
OperatorPrivateKey = ""
FundingPrivateKey = ""
SealPrivateKey = ""
ApprovalPrivateKey = ""
# block syncer configuration
[BlockSyncerCfg]
Modules = ["epoch", "bucket", "object", "payment"]
Dsn = "localhost:3308"
# p2p node configuration
[P2PConfig]
ListenAddress = ""
PrivKey = ""
Bootstrap = []
PingPeriod = 0
# log configuration
[LogCfg]
Level = "info"
Path = "./gnfd-sp.log"
```

### Start

```shell
# start sp
./gnfd-sp --config ${config_file_path}
```

### Add Greenfield Chain
[Add SP to Greenfield](https://github.com/bnb-chain/greenfield/blob/master/docs/cli/storage-provider.md)

## Document
* [Greenfield Whitepaper](https://github.com/bnb-chain/greenfield-whitepaper): the official Greenfield Whitepaper. 
* [Greenfield](https://github.com/bnb-chain/greenfield#readme): the Greenfield documents.
* [SP Introduce](docs/readme.md): the Greenfield Storage Provider documents.
* [Storage Metadata](https://github.com/bnb-chain/greenfield/blob/master/docs/core-concept/storage-metadata-models.md) : the storage metadata on Greenfield Chain.
* [SP on Greenfield](https://github.com/bnb-chain/greenfield/blob/master/docs/modules/storage_provider_management.md): the SP module on Greenfield Chain.
* [Data Availability Challenge](https://github.com/bnb-chain/greenfield/blob/master/docs/modules/data_availability_challenge.md): the correctness of payload be stored in SP. 
* [SP Deployment](docs/tutorial/01-deployment.md): the detailed introduction to deploying sp.
* [SP Local Setup](docs/run-book/03-local.toml): the introduction to set up local SP env for testing.

## Related Projects
* [Greenfield](https://github.com/bnb-chain/greenfield): the Golang implementation of the Greenfield Blockchain.
* [Greenfield-Common](https://github.com/bnb-chain/greenfield-common): the Greenfield common package.
* [Reed-Solomon](https://github.com/klauspost/reedsolomon): the Reed-Solomon Erasure package in prue Go, with speeds exceeding 1GB/s/cpu core.
* [Juno](https://github.com/bnb-chain/juno): the Cosmos Hub blockchain data aggregator and exporter package.
* [Greenfield-Go-SDK](https://github.com/bnb-chain/greenfield-go-sdk): the Greenfield SDK, interact with SP, Greenfield and Tendermint.


## Contribution
Thank you for considering to help out with the source code! We welcome contributions from 
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to Greenfield Storage Provider, please fork, fix, commit and 
send a pull request for the maintainers to review and merge into the main code base. 
If you wish to submit more complex changes though, please check up with the core devs first 
through github issue(going to have a discord channel soon) to ensure those changes are in 
line with the general philosophy of the project and/or get some early feedback which can make 
both your efforts much lighter as well as our review and merge procedures quick and simple.

## Licence (pending)
