# Greenfield Storage Provider

Greenfield Storage Provider (abbreviated SP) is storage service infrastructure provider. It uses Greenfield as the ledger and the single source of truth. Each SP can and will respond to users' requests to write (upload) and read (download) data, and serve as the gatekeeper for user rights and authentications.

## Disclaimer

**The software and related documentation are under active development, all subject to potential future change without notification and not ready for production use. The code and security audit have not been fully completed and not ready for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## SP Core

SPs store the objects' real data, i.e. the payload data. Each SP runs its own object storage system. Similar to Amazon S3 and other object store systems, the objects stored on SPs are immutable. The users may delete and re-create the object (under the different ID, or under the same ID after certain publicly declared settings), but they cannot modify it.

SPs have to register themselves first by depositing on the Greenfield blockchain as their "Service Stake". Greenfield validators will go through a dedicated governance procedure to vote for the SPs of their election. SPs are encouraged to advertise their information and prove to the community their capability, as SPs have to provide a professional storage system with high-quality SLA.

SPs provide publicly accessible APIs for users to upload, download, and manage data. These APIs are very similar to Amazon S3 APIs so that existing developers may feel familiar enough to write code for it. Meanwhile, they provide each other REST APIs and form another white-listed P2P network to communicate with each other to ensure data availability and redundancy. There will also be a P2P-based upload/download network across SPs and user-end client software to facilitate easy connections and fast data download, which is similar to BitTorrent.

Among the multiple SPs that one object is stored on, one SP will be the "Primary SP", while the others are "Secondary SP".

When users want to write an object into Greenfield, they or the client software they use must specify the primary SP. Primary SP should be used as the only SP to download the data. Users can change the primary SP for their objects later if they are not satisfied with its service.

## Quick Started

*Note*: Requires [Go 1.20+](https://go.dev/dl/)

## Compile SP

Compilation dependencies:

- [Golang](https://go.dev/dl): SP is written in Golang, you need to install it. Golang version requires `1.20+`.
- [Buf](https://buf.build/docs/installation/): A new way of working with Protocol Buffers. SP uses Buf to manage proto files.
- [protoc-gen-gocosmos](https://github.com/cosmos/gogoproto): Protocol Buffers for Go with Gadgets. SP uses this protobuf compiler to generate pb.go files.
- [mockgen](https://github.com/uber-go/mock): A mocking framework for the Go programming language that is used in unit test.
- [jq](https://stedolan.github.io/jq/): Command-line JSON processor. Users should install jq according to your operating system.

```shell
# clone source code
git clone https://github.com/bnb-chain/greenfield-storage-provider.git

cd greenfield-storage-provider

# install dependent tools: buf, protoc-gen-gocosmos and mockgen
make install-tools

# compile sp
make build

# move to build directory
cd build

# execute gnfd-sp binary file
./gnfd-sp version

# show the gnfd-sp version information
Greenfield Storage Provider
    __                                                       _     __
    _____/ /_____  _________ _____ ____     ____  _________ _   __(_)___/ /__  _____
    / ___/ __/ __ \/ ___/ __  / __  / _ \   / __ \/ ___/ __ \ | / / / __  / _ \/ ___/
    (__  ) /_/ /_/ / /  / /_/ / /_/ /  __/  / /_/ / /  / /_/ / |/ / / /_/ /  __/ /
    /____/\__/\____/_/   \__,_/\__, /\___/  / .___/_/   \____/|___/_/\__,_/\___/_/
    /____/       /_/

Version : vx.x.x
Branch  : master
Commit  : 342930b89466c15653af2f3695cfc72f6466d4b8
Build   : go1.20.3 darwin arm64 2023-06-20 10:31

# show the gnfd-sp help info
./gnfd-sp -h
```

### Note

If you've already executed `make install-tools` command in your shell, but you failed to make build and encountered one of the following error messages:

```shell
# error message 1
buf: command not found
# you can execute the following command, assumed that you installed golang in /usr/local/go/bin. Other OS are similar.
GO111MODULE=on GOBIN=/usr/local/go/bin go install github.com/bufbuild/buf/cmd/buf@v1.25.0

# error message 2
Failure: plugin gocosmos: could not find protoc plugin for name gocosmos - please make sure protoc-gen-gocosmos is installed and present on your $PATH
# you can execute the fowllowing command, assumed that you installed golang in /usr/local/go/bin. Other OS are similar.
GO111MODULE=on GOBIN=/usr/local/go/bin go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest

# if you want to execute unit test of sp, you should execute the following command, assumed that you installed golang in /usr/local/go/bin. Other OS are similar.
GO111MODULE=on GOBIN=/usr/local/go/bin go install go.uber.org/mock/mockgen@latest
```

Above error messages are due to users don't set go env correctly. More info users can search `GOROOT`, `GOPATH` and `GOBIN`.

## SP Dependencies

If a user wants to start SP in local mode or testnet mode, you must prepare `SPDB`, `BSDB` and `PieceStore` dependencies.

### SPDB and BSDB

SP uses [SPDB](../modules/spdb.md) and [BSDB](../modules/bsdb.md) to store some metadata such as object info, object integrity hash, etc. These two DBs now use `RDBMS` to complete corresponding function.

Users now can use `MySQL` or `MariaDB` to store metadata.The following lists the supported RDBMS:

1. [MySQL](https://www.mysql.com/)
2. [MariaDB](https://mariadb.org/)

More types of database such as `PostgreSQL` or NewSQL will be supported in the future.

### PieceStore

Greenfield is a decentralized data storage system which uses object storage as the main data storage system. SP encapsulates data storage as [PieceStore](../modules/piece-store.md) which provides common interfaces to be compatible with multiple data storage systems. Therefore, if a user wants to join SP or test the function of SP, you must use a data storage system.

The following lists the supported data storage systems:

1. [AWS S3](https://aws.amazon.com/s3/): An object storage can be used in production environment.
2. [MinIO](https://min.io/): An object storage can be used in production environment which is compatible with AWS S3.
3. [POSIX Filesystem](https://en.wikipedia.org/wiki/POSIX): Local filesystem is used for experiencing the basic features of SP and understanding how SP works. The piece data created by SP cannot be got within the network and can only be used on a single machine.

### Install Dependencies

#### Install MySQL in CentOS

1. Install MySQL yum package

```shell
# 1. Download MySQL yum package
wget http://repo.mysql.com/mysql57-community-release-el7-10.noarch.rpm

# 2. Install MySQL source
rpm -Uvh mysql57-community-release-el7-10.noarch.rpm

# 3. Install public key
rpm --import https://repo.mysql.com/RPM-GPG-KEY-mysql-2022

# 4. Install MySQL server
yum install -y mysql-community-server

# 5. Start MySQL
systemctl start mysqld.service

# 6. Check whether the startup is successful
systemctl status mysqld.service

# 7. Get temporary password
grep 'temporary password' /var/log/mysqld.log 

# 8. Login MySQL through temporary password
# After you log in with the temporary password, do not perform any other operations. Otherwise, an error will occur. In this case, you need to change the password
mysql -uroot -p

# 9. change MySQL password rules
mysql> set global validate_password_policy=0;
mysql> set global validate_password_length=1;
mysql> ALTER USER 'root'@'localhost' IDENTIFIED BY 'yourpassword';
```


### Configuration

#### Make configuration template

```shell
# dump default configuration
./gnfd-sp config.dump
```

```toml
# optional
Env = ''
# optional
AppID = ''
# optional
Server = []
# optional
GRPCAddress = ''

[SpDB]
# required
User = ''
# required
Passwd = ''
# required
Address = ''
# required
Database = ''
# optional
ConnMaxLifetime = 0
# optional
ConnMaxIdleTime = 0
# optional
MaxIdleConns = 0
# optional
MaxOpenConns = 0

[BsDB]
# required
User = ''
# required
Passwd = ''
# required
Address = ''
# required
Database = ''
# optional
ConnMaxLifetime = 0
# optional
ConnMaxIdleTime = 0
# optional
MaxIdleConns = 0
# optional
MaxOpenConns = 0

[PieceStore]
# required
Shards = 0

[PieceStore.Store]
# required
Storage = ''
# optional
BucketURL = ''
# optional
MaxRetries = 0
# optional
MinRetryDelay = 0
# optional
TLSInsecureSkipVerify = false
# required
IAMType = ''

[Chain]
# required
ChainID = ''
# required
ChainAddress = []
# optional
SealGasLimit = 0
# optional
SealFeeAmount = 0
# optional
RejectSealGasLimit = 0
# optional
RejectSealFeeAmount = 0
# optional
DiscontinueBucketGasLimit = 0
# optional
DiscontinueBucketFeeAmount = 0
# optional
CreateGlobalVirtualGroupGasLimit = 0
# optional
CreateGlobalVirtualGroupFeeAmount = 0
# optional
CompleteMigrateBucketGasLimit = 0
# optional
CompleteMigrateBucketFeeAmount = 0

[SpAccount]
# required
SpOperatorAddress = ''
# required
OperatorPrivateKey = ''
# required
SealPrivateKey = ''
# required
ApprovalPrivateKey = ''
# required
GcPrivateKey = ''
# required
BlsPrivateKey = ''

[Endpoint]
# required
ApproverEndpoint = ''
# required
ManagerEndpoint = ''
# required
DownloaderEndpoint = ''
# required
ReceiverEndpoint = ''
# required
MetadataEndpoint = ''
# required
UploaderEndpoint = ''
# required
P2PEndpoint = ''
# required
SignerEndpoint = ''
# required
AuthenticatorEndpoint = ''

[Approval]
# optional
BucketApprovalTimeoutHeight = 0
# optional
ObjectApprovalTimeoutHeight = 0
# optional
ReplicatePieceTimeoutHeight = 0

[Bucket]
# optional
AccountBucketNumber = 0
# optional
FreeQuotaPerBucket = 0
# optional
MaxListReadQuotaNumber = 0
# optional
MaxPayloadSize = 0

[Gateway]
# required
DomainName = ''
# required
HTTPAddress = ''

[Executor]
# optional
MaxExecuteNumber = 0
# optional
AskTaskInterval = 0
# optional
AskReplicateApprovalTimeout = 0
# optional
AskReplicateApprovalExFactor = 0.0
# optional
ListenSealTimeoutHeight = 0
# optional
ListenSealRetryTimeout = 0
# optional
MaxListenSealRetry = 0

[P2P]
# optional
P2PPrivateKey = ''
# optional
P2PAddress = ''
# optional
P2PAntAddress = ''
# optional
P2PBootstrap = []
# optional
P2PPingPeriod = 0

[Parallel]
# optional
GlobalCreateBucketApprovalParallel = 0
# optional
GlobalCreateObjectApprovalParallel = 0
# optional
GlobalMaxUploadingParallel = 0
# optional
GlobalUploadObjectParallel = 0
# optional
GlobalReplicatePieceParallel = 0
# optional
GlobalSealObjectParallel = 0
# optional
GlobalReceiveObjectParallel = 0
# optional
GlobalGCObjectParallel = 0
# optional
GlobalGCZombieParallel = 0
# optional
GlobalGCMetaParallel = 0
# optional
GlobalRecoveryPieceParallel = 0
# optional
GlobalMigrateGVGParallel = 0
# optional
GlobalBackupTaskParallel = 0
# optional
GlobalDownloadObjectTaskCacheSize = 0
# optional
GlobalChallengePieceTaskCacheSize = 0
# optional
GlobalBatchGcObjectTimeInterval = 0
# optional
GlobalGcObjectBlockInterval = 0
# optional
GlobalGcObjectSafeBlockDistance = 0
# optional
GlobalSyncConsensusInfoInterval = 0
# optional
UploadObjectParallelPerNode = 0
# optional
ReceivePieceParallelPerNode = 0
# optional
DownloadObjectParallelPerNode = 0
# optional
ChallengePieceParallelPerNode = 0
# optional
AskReplicateApprovalParallelPerNode = 0
# optional
QuerySPParallelPerNode = 0
# required
DiscontinueBucketEnabled = false
# optional
DiscontinueBucketTimeInterval = 0
# required
DiscontinueBucketKeepAliveDays = 0
# optional
LoadReplicateTimeout = 0
# optional
LoadSealTimeout = 0

[Task]
# optional
UploadTaskSpeed = 0
# optional
DownloadTaskSpeed = 0
# optional
ReplicateTaskSpeed = 0
# optional
ReceiveTaskSpeed = 0
# optional
SealObjectTaskTimeout = 0
# optional
GcObjectTaskTimeout = 0
# optional
GcZombieTaskTimeout = 0
# optional
GcMetaTaskTimeout = 0
# optional
SealObjectTaskRetry = 0
# optional
ReplicateTaskRetry = 0
# optional
ReceiveConfirmTaskRetry = 0
# optional
GcObjectTaskRetry = 0
# optional
GcZombieTaskRetry = 0
# optional
GcMetaTaskRetry = 0

[Monitor]
# required
DisableMetrics = false
# required
DisablePProf = false
# required
DisableProbe = false
# required
MetricsHTTPAddress = ''
# required
PProfHTTPAddress = ''
# required
ProbeHTTPAddress = ''

[Rcmgr]
# optional
DisableRcmgr = false

[Log]
# optional
Level = ''
# optional
Path = ''

[Metadata]
# required
IsMasterDB = false
# optional
BsDBSwitchCheckIntervalSec = 0

[BlockSyncer]
# required
Modules = []
# optional
BsDBWriteAddress = ''
# required
Workers = 0

[APIRateLimiter]
# every line should represent one entry of gateway route. The comment after each line must contain which route name it represents.
# Most of APIs has a qps number, offered by QA team.  That usually means the max qps for the whole 4 gateway cluster.
# How to setup the RateLimit value, it is a sophistcated question and need take a lot of factors into account.
# 1. For most query-APIs, we can setup a rate limit up to the 1/4 of max qps, as the config is for only one gateway instance.
# 2. Also we avoid to setup a too large or too small rate limit value.
# 3. For upload/download APIs, it is diffiult to use a rate limit as a protect mechanism for the servers. Because the performance of upload/download interactions usually dependens on how large the file is processed.
# 4. We tetatively setup 50~75 as the rate limit for the download/upload APIs and we can ajdust them once we have a better experience.
# 5. The rate limt config will upgraded in next version to use http methods and virtual-host/path style as part of the matching keys.

# optional
PathPattern = [
    {Key = "/auth/request_nonce", Method = "GET", Names = ["GetRequestNonce"]}, 
    {Key = "/auth/update_key", Method = "POST", Names = ["UpdateUserPublicKey"]}, 
    {Key = "/permission/.+/[^/]*/.+", Method = "GET", Names = ["VerifyPermission"]},
    {Key = "/greenfield/admin/v1/get-approval", Method = "GET", Names = ["GetApproval"]},
    {Key = "/greenfield/admin/v1/challenge", Method = "GET", Names = ["GetChallengeInfo"]},
    {Key = "/greenfield/receiver/v1/replicate-piece", Method = "PUT", Names = ["ReplicateObjectPiece"]},
    {Key = "/greenfield/recovery/v1/get-piece", Method = "GET", Names = ["RecoveryPiece"]},
    {Key = "/greenfield/migrate/v1/notify-migrate-swap-out-task", Method = "POST", Names = ["NotifyMigrateSwapOut"]},
    {Key = "/greenfield/migrate/v1/migrate-piece", Method = "GET", Names = ["MigratePiece"]},
    {Key = "/greenfield/migrate/v1/migration-bucket-approval", Method = "GET", Names = ["MigrationBucketApproval"]},
    {Key = "/greenfield/migrate/v1/get-swap-out-approval", Method = "GET", Names = ["SwapOutApproval"]},
    {Key = "/download/[^/]*/.+", Method = "GET", Names = ["DownloadObjectByUniversalEndpoint"]},{Key = "/download", Method = "GET", Names = ["DownloadObjectByUniversalEndpoint"]},
    {Key = "/view/[^/]*/.+", Method = "GET", Names = ["ViewObjectByUniversalEndpoint"]},{Key = "/view", Method = "GET", Names = ["ViewObjectByUniversalEndpoint"]},
    {Key = "/status", Method = "GET", Names = ["GetStatus"]},
    {Key = "/.+/.+[?]offset.*", Method = "POST", Names = ["ResumablePutObject"]},
    {Key = "/.+/.+[?]upload-context.*", Method = "GET", Names = ["QueryResumeOffset"]},
    {Key = "/.+/.+[?]upload-progress.*", Method = "GET", Names = ["QueryUploadProgress"]},
    {Key = "/.+/.+[?]bucket-meta.*", Method = "GET", Names = ["GetBucketMeta"]},
    {Key = "/.+/.+[?]object-meta.*", Method = "GET", Names = ["GetObjectMeta"]},
    {Key = "/.+/.+[?]object-policies.*", Method = "GET", Names = ["ListObjectPolicies"]},
    {Key = "/.+[?]read-quota.*", Method = "GET", Names = ["GetBucketReadQuota"]},
    {Key = "/.+[?]list-read-quota.*", Method = "GET", Names = ["listBucketReadRecord"]},
    {Key = "/[?].*group-query.*", Method = "GET", Names = ["getGroupList"]},
    {Key = "/[?].*objects-query.*", Method = "GET", Names = ["listObjectsByIDs"]},
    {Key = "/[?].*buckets-query.*", Method = "GET", Names = ["listBucketsByIDs"]},
    {Key = "/[?].*verify-id.*", Method = "GET", Names = ["verifyPermissionByID"]},
    {Key = "/[?].*user-groups.*", Method = "GET", Names = ["getUserGroups"]},
    {Key = "/[?].*group-members.*", Method = "GET", Names = ["getGroupMembers"]},
    {Key = "/[?].*owned-groups.*", Method = "GET", Names = ["getUserOwnedGroups"]},
    
    {Key = "/.+/$", Method = "GET", Names = ["ListObjectsByBucket"]},
    {Key = "/.+/.+", Method = "GET", Names = ["ListObjectsByBucket"]},
    {Key = "/.+/.+", Method = "PUT", Names = ["PutObject"]},
    {Key = "/$", Method = "GET", Names = ["GetUserBuckets"]},

]

NameToLimit = [
    {Name = "GetRequestNonce", RateLimit = 100, RatePeriod = 'S'}, # requestNonceRouterName 3000qps
    {Name = "UpdateUserPublicKey", RateLimit = 100, RatePeriod = 'S'}, # updateUserPublicKeyRouterName 4000qps
    {Name = "VerifyPermission", RateLimit = 100, RatePeriod = 'S'}, # verifyPermissionRouterName  1200qps
    {Name = "GetApproval", RateLimit = 35, RatePeriod = 'S'}, # approvalRouterName  150qps
    {Name = "GetChallengeInfo", RateLimit = 20, RatePeriod = 'S'}, # getChallengeInfoRouterName, no test data
    {Name = "ReplicateObjectPiece", RateLimit = 1000, RatePeriod = 'S'},  # replicateObjectPieceRouterName, no test data. Internal API among sps, no rate limit is needed.
    {Name = "RecoveryPiece", RateLimit = 1000, RatePeriod = 'S'}, # recoveryPieceRouterName, no test data. Internal API among sps, no rate limit is needed.
    {Name = "NotifyMigrateSwapOut", RateLimit = 10, RatePeriod = 'S'},  # notifyMigrateSwapOutRouterName, no test data. Internal API among sps, no rate limit is needed.
    {Name = "MigratePiece", RateLimit = 10, RatePeriod = 'S'}, # migratePieceRouterName, no test data
    {Name = "MigrationBucketApproval", RateLimit = 10, RatePeriod = 'S'}, # migrationBucketApprovalName, no test data
    {Name = "SwapOutApproval", RateLimit = 10, RatePeriod = 'S'}, # swapOutApprovalName, no test data
    {Name = "DownloadObjectByUniversalEndpoint", RateLimit = 50, RatePeriod = 'S'}, # downloadObjectByUniversalEndpointName, 50qps
    {Name = "ViewObjectByUniversalEndpoint", RateLimit = 50, RatePeriod = 'S'}, # viewObjectByUniversalEndpointName, 50qps
    {Name = "GetStatus", RateLimit = 200, RatePeriod = 'S'},# getStatusRouterName, 2000qps
    {Name = "ResumablePutObject", RateLimit = 30, RatePeriod = 'S'}, # resumablePutObjectRouterName , test data is same as putObject object 10qps
    {Name = "QueryResumeOffset", RateLimit = 30, RatePeriod = 'S'},  # queryResumeOffsetName, test data is same as putObject object 10qps
    {Name = "QueryUploadProgress", RateLimit = 50, RatePeriod = 'S'}, # queryUploadProgressRouterName, test data is same as putObject object 10qps
    {Name = "GetBucketMeta", RateLimit = 100, RatePeriod = 'S'}, # getBucketMetaRouterName, 400qps
    {Name = "GetObjectMeta", RateLimit = 100, RatePeriod = 'S'}, # getObjectMetaRouterName, 400qps
    {Name = "ListObjectPolicies", RateLimit = 200, RatePeriod = 'S'}, # listObjectPoliciesRouterName, 2000qps
    {Name = "GetBucketReadQuota", RateLimit = 200, RatePeriod = 'S'}, # getBucketReadQuotaRouterName
    {Name = "ListBucketReadRecord", RateLimit = 100, RatePeriod = 'S'}, # listBucketReadRecordRouterName
    {Name = "GetGroupList", RateLimit = 200, RatePeriod = 'S'}, # getGroupListRouterName， similar to getUserGroupsRouterName, 2000qps
    {Name = "ListObjectsByIDs", RateLimit = 200, RatePeriod = 'S'}, # listObjectsByIDsRouterName, 1200qps
    {Name = "ListBucketsByIDs", RateLimit = 200, RatePeriod = 'S'}, # listBucketsByIDsRouterName, 2000qps
    {Name = "VerifyPermissionByID", RateLimit = 200, RatePeriod = 'S'}, # verifyPermissionByIDRouterName, 1200qps
    {Name = "GetUserGroups", RateLimit = 200, RatePeriod = 'S'}, # getUserGroupsRouterName, 2000qps
    {Name = "GetGroupMembers", RateLimit = 200, RatePeriod = 'S'}, # getGroupMembersRouterName, 2000qps
    {Name = "GetUserOwnedGroups", RateLimit = 200, RatePeriod = 'S'}, # getUserOwnedGroupsRouterName, 2000qps
    
    {Name = "ListObjectsByBucket", RateLimit = 75, RatePeriod = 'S'}, # listObjectsByBucketRouterName, 300qps
    {Name = "GetObject", RateLimit = 75, RatePeriod = 'S'}, # getObjectRouterName, 100 qps
    {Name = "PutObject", RateLimit = 75, RatePeriod = 'S'}, # putObjectRouterName, 100 qps
    {Name = "GetUserBuckets", RateLimit = 75, RatePeriod = 'S'}] # getUserBucketsRouterName, 1000 qps

# optional
HostPattern = []
# optional
APILimits = []

[APIRateLimiter.IPLimitCfg]
# optional
On = false
# optional
RateLimit = 0
# optional
RatePeriod = ''

[Manager]
# optional
EnableLoadTask = false
# optional
SubscribeSPExitEventIntervalSec = 0
# optional
SubscribeSwapOutExitEventIntervalSec = 0
# optional
SubscribeBucketMigrateEventIntervalSec = 0
# optional
GVGPreferSPList = []
# optional
SPBlackList = []
```

## App info
These fields are optional and they can.
```
GRPCAddress = '0.0.0.0:9333'
```
## Database

To config `[SpDB]`, `[BsDB]`, you have to input the `user name`, `db password`,`db address`  and  `db name` in these fields.

## PieceStore

To config `[PieceStore]` and `[PieceStore.Store]`, you can read the details in this [doc](./piece-store.md)

## Chain info

* `ChainID` of testnet is `greenfield_5600-1`.
* `ChainAddress` is RPC endpoint of testnet, you can find RPC info [here](../../../api/endpoints.md)

## SpAccount
These private keys are generated during wallet setup.


## Endpoint
`[Endpoint]` specified the URL of different services.

For single-machine host (not recommended):
```
[Endpoint]
ApproverEndpoint = ''
ManagerEndpoint = ''
DownloaderEndpoint = ''
ReceiverEndpoint = ''
MetadataEndpoint = ''
UploaderEndpoint = ''
P2PEndpoint = ''
SignerEndpoint = ''
AuthenticatorEndpoint = ''
```

For K8S cluster:
```
[Endpoint]
ApproverEndpoint = 'manager:9333'
ManagerEndpoint = 'manager:9333'
DownloaderEndpoint = 'downloader:9333'
ReceiverEndpoint = 'receiver:9333'
MetadataEndpoint = 'metadata:9333'
UploaderEndpoint = 'uploader:9333'
P2PEndpoint = 'p2p:9333'
SignerEndpoint = 'signer:9333'
AuthenticatorEndpoint = 'localhost:9333'
```

## P2P
* `P2PPrivateKey` and `node_id` is generated by `./gnfd-sp p2p.create.key -n 1`

* `P2PAntAddress` is your load balance address. If you don't have a load balance address, you should have a public IP and use it in `P2PAddress`. It consists of `ip:port`.

* `P2PBootstrap` can be left empty.

## Gateway
```
[Gateway]
DomainName = 'region.sp-name.com'
```
The correct configuration should not include the protocol prefix `https://`.

## BlockSyncer
Here is block_syncer config.
The configuration of BsDBWriteAddress can be the same as the BSDB.Address module here. To enhance performance, you can set up the write database address here and the corresponding read database address in BSDB.
```
Modules = ['epoch','bucket','object','payment','group','permission','storage_provider','prefix_tree', 'virtual_group','sp_exit_events','object_id_map','general']
BsDBWriteAddress = 'localhost:3306'
Workers = 50
```

### Start

```shell
# start sp
./gnfd-sp --config ${config_file_path}
```

### Join Greenfield Testnet

[Run Testnet SP Node](https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/run-book/run-testnet-SP-node)

## Document

* [Greenfield Whitepaper](https://github.com/bnb-chain/greenfield-whitepaper): The official Greenfield Whitepaper.
* [Greenfield](https://docs.bnbchain.org/greenfield-docs/docs/guide/greenfield-blockchain/overview): The Greenfield documents.
* [Storage Module on Greenfield](https://docs.bnbchain.org/greenfield-docs/docs/guide/greenfield-blockchain/modules/storage-module): The storage module on Greenfield Chain.
* [Storage Provider on Greenfield](https://docs.bnbchain.org/greenfield-docs/docs/guide/greenfield-blockchain/modules/storage-provider): The storage provider on Greenfield Chain.
* [Data Availability Challenge](https://docs.bnbchain.org/greenfield-docs/docs/guide/greenfield-blockchain/modules/data-availability-challenge): The correctness of payload be stored in SP.
* [SP Introduction](https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/introduction/overview): The Greenfield Storage Provider documents.
* [SP Compiling and Dependencies](https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/run-book/compile-dependences): The detailed introduction to sp compiling and dependencies.
* [Run Local SP Network](https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/run-book/run-local-SP-network): The introduction to run local SP env for testing.
* [Run Testnet SP Node](https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/run-book/run-testnet-SP-node): The introduction to run testnet SP node.

## Related Projects

* [Greenfield](https://github.com/bnb-chain/greenfield): The Golang implementation of the Greenfield Blockchain.
* [Greenfield-Go-SDK](https://github.com/bnb-chain/greenfield-go-sdk): The Greenfield SDK, interact with SP, Greenfield and Tendermint.
* [Greenfield Cmd](https://github.com/bnb-chain/greenfield-cmd): Greenfield client cmd tool, supporting commands to make requests to greenfield.
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
