# Greenfield Storage Provider

Greenfield Storage Provider (abbreviated SP) is storage service infrastructure provider. It uses Greenfield as the ledger and the single source of truth. Each SP can and will respond to users' requests to write (upload) and read (download) data, and serve as the gatekeeper for user rights and authentications.

## Disclaimer

**The software and related documentation are under active development, all subject to potential future change withoutnotification and not ready for production use. The code and security audit have not been fully completed and not ready for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## SP Core

SPs store the objects' real data, i.e. the payload data. Each SP runs its own object storage system. Similar to Amazon S3 and other object store systems, the objects stored on SPs are immutable. The users may delete and re-create the object (under the different ID, or under the same ID after certain publicly declared settings), but they cannot modify it.

SPs have to register themselves first by depositing on the Greenfield blockchain as their "Service Stake". Greenfield validators will go through a dedicated governance procedure to vote for the SPs of their election. SPs are encouraged to advertise their information and prove to the community their capability, as SPs have to provide a professional storage system with high-quality SLA.

SPs provide publicly accessible APIs for users to upload, download, and manage data. These APIs are very similar to Amazon S3 APIs so that existing developers may feel familiar enough to write code for it. Meanwhile, they provide each other REST APIs and form another white-listed P2P network to communicate with each other to ensure data availability and redundancy. There will also be a P2P-based upload/download network across SPs and user-end client software to facilitate easy connections and fast data download, which is similar to BitTorrent.

Among the multiple SPs that one object is stored on, one SP will be the "Primary SP", while the others are "Secondary SP".

When users want to write an object into Greenfield, they or the client software they use must specify the primary SP. Primary SP should be used as the only SP to download the data. Users can change the primary SP for their objects later if they are not satisfied with its service.

## Quick Started

*Note*: Requires [Go 1.20+](https://go.dev/dl/)

### Compile SP

For detailed information about compiling SP, you can refer this doc to [Compile SP](https://greenfield.bnbchain.org/docs/guide/storage-provider/run-book/compile-dependences.html#compile-sp).

```shell
# install tools
make install-tools

# build gnfd-sp
make build && cd build 
# show version
./gnfd-sp version

Greenfield Storage Provider
    __                                                       _     __
    _____/ /_____  _________ _____ ____     ____  _________ _   __(_)___/ /__  _____
    / ___/ __/ __ \/ ___/ __  / __  / _ \   / __ \/ ___/ __ \ | / / / __  / _ \/ ___/
    (__  ) /_/ /_/ / /  / /_/ / /_/ /  __/  / /_/ / /  / /_/ / |/ / / /_/ /  __/ /
    /____/\__/\____/_/   \__,_/\__, /\___/  / .___/_/   \____/|___/_/\__,_/\___/_/
    /____/       /_/
Version : vx.x.x
Branch  : master
Commit  : bfc32b9748c11d74493f93c420744ade4dbc18ac
Build   : go1.20.3 darwin arm64 2023-05-12 13:37

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
Server = []
GrpcAddress = '0.0.0.0:9333'

[SpDB]
User = '${db_user}'
Passwd = '${db_password}'
Address = '${db_address}'
Database = 'storage_provider_db'

[BsDB]
User = '${db_user}'
Passwd = '${db_password}'
Address = '${db_address}'
Database = 'block_syncer'

[BsDBBackup]
User = '${db_user}'
Passwd = '${db_password}'
Address = '${db_address}'
Database = 'block_syncer_backup'

[PieceStore]
Shards = 0

[PieceStore.Store]
Storage = 's3'
BucketURL = '${bucket_url}'
MaxRetries = 5
MinRetryDelay = 0
TLSInsecureSkipVerify = false
IAMType = 'SA'

[Chain]
ChainID = '${chain_id}'
ChainAddress = ['${chain_address}']

[SpAccount]
SpOperateAddress = '${sp_operator_address}'
OperatorPrivateKey = '${operator_private_key}'
FundingPrivateKey = '${funding_private_key}'
SealPrivateKey = '${seal_private_key}'
ApprovalPrivateKey = '${approval_private_key}'
GcPrivateKey = '${gc_private_key}'

[Endpoint]
ApproverEndpoint = 'approver:9333'
ManagerEndpoint = 'manager:9333'
DownloaderEndpoint = 'downloader:9333'
ReceiverEndpoint = 'receiver:9333'
MetadataEndpoint = 'metadata:9333'
UploaderEndpoint = 'uploader:9333'
P2PEndpoint = 'p2p:9333'
SignerEndpoint = 'signer:9333'
AuthorizerEndpoint = 'localhost:9333'

[Gateway]
Domain = '${gateway_domain_name}'
HttpAddress = '0.0.0.0:9033'

[P2P]
P2PPrivateKey = '${p2p_private_key}'
P2PAddress = '0.0.0.0:9933'
P2PAntAddress = '${p2p_ant_address}'
P2PBootstrap = ['${node_id@p2p_ant_address}']
P2PPingPeriod = 0

[Parallel]
DiscontinueBucketEnabled = true
DiscontinueBucketKeepAliveDays = 2

[Monitor]
DisableMetrics = false
DisablePProf = false
MetricsHttpAddress = '0.0.0.0:24367'
PProfHttpAddress = '0.0.0.0:24368'

[Rcmgr]
DisableRcmgr = false

[Metadata]
IsMasterDB = true
BsDBSwitchCheckIntervalSec = 30

[BlockSyncer]
Modules = ['epoch','bucket','object','payment','group','permission','storage_provider','prefix_tree']
Dsn = '${dsn}'
DsnSwitched = ''
RecreateTables = false
Workers = 50
EnableDualDB = false

[APIRateLimiter]
PathPattern = [{Key = ".*request_nonc.*", RateLimit = 10, RatePeriod = 'S'},{Key = ".*1l65v.*", RateLimit = 20, RatePeriod = 'S'}]
HostPattern = [{Key = ".*vfdxy.*", RateLimit = 15, RatePeriod = 'S'}]

[APIRateLimiter.IPLimitCfg]
On = true
RateLimit = 5000
RatePeriod = 'S'
```

### Start

```shell
# start sp
./gnfd-sp --config ${config_file_path}
```

### Add Greenfield Chain

[Add SP to Greenfield](https://github.com/bnb-chain/greenfield-docs/blob/master/src/guide/storage-provider/run-book/run-testnet-SP-node.md)

## Document

* [Greenfield Whitepaper](https://github.com/bnb-chain/greenfield-whitepaper): The official Greenfield Whitepaper.
* [Greenfield](https://greenfield.bnbchain.org/docs/guide/greenfield-blockchain/overview.html): The Greenfield documents.
* [Storage Module on Greenfield](https://greenfield.bnbchain.org/docs/guide/greenfield-blockchain/modules/storage-module.html): The storage module on Greenfield Chain.
* [Storage Provider on Greenfield](https://greenfield.bnbchain.org/docs/guide/greenfield-blockchain/modules/storage-provider.html): The storage provider on Greenfield Chain.
* [Data Availability Challenge](https://greenfield.bnbchain.org/docs/guide/greenfield-blockchain/modules/data-availability-challenge.html): The correctness of payload be stored in SP.
* [SP Introduction](https://greenfield.bnbchain.org/docs/guide/storage-provider/introduction/overview.html): The Greenfield Storage Provider documents.
* [SP Compiling and Dependencies](https://greenfield.bnbchain.org/docs/guide/storage-provider/run-book/compile-dependences.html): The detailed introduction to sp compiling and dependencies.
* [Run Local SP Network](https://greenfield.bnbchain.org/docs/guide/storage-provider/run-book/run-local-SP-network.html): The introduction to run local SP env for testing.
* [Run Testnet SP Node](https://greenfield.bnbchain.org/docs/guide/storage-provider/run-book/run-testnet-SP-node.html): The introduction to run testnet SP node.

## Related Projects

* [Greenfield](https://github.com/bnb-chain/greenfield): The Golang implementation of the Greenfield Blockchain.
* [Greenfield-Go-SDK](https://github.com/bnb-chain/greenfield-go-sdk): The Greenfield SDK, interact with SP, Greenfield and Tendermint.
* [Greenfield-Common](https://github.com/bnb-chain/greenfield-common): The Greenfield common package.
* [Reed-Solomon](https://github.com/klauspost/reedsolomon): The Reed-Solomon Erasure package in prue Go, with speeds exceeding 1GB/s/cpu core.
* [Juno](https://github.com/bnb-chain/juno): The Cosmos Hub blockchain data aggregator and exporter package.

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from 
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to Greenfield Storage Provider, please fork, fix, commit and 
send a pull request for the maintainers to review and merge into the main code base. 
If you wish to submit more complex changes though, please check up with the core devs first 
through github issue(going to have a discord channel soon) to ensure those changes are in 
line with the general philosophy of the project and/or get some early feedback which can make 
both your efforts much lighter as well as our review and merge procedures quick and simple.

## License

The greenfield storage provider library (i.e. all code outside the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The greenfield storage provider binaries (i.e. all code inside the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
