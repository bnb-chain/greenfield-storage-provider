# Module

The Module is a self-contained logical component of SP, with necessary interactions between modules handled by the GfSp
framework. The implementation of the module can be customized as needed. For instance, while the GfSp framework requires
object approval before uploading, SP can customize whether to agree with the approval.

## Concept

### Front Modules

The Front Modules are responsible for handling user requests. The Gater generates corresponding tasks and sends them to
the Front Modules. The Front Modules verify the correctness of the request and perform additional tasks after handling
the request. To accomplish this, the Front Modules have three interfaces for each task type: `PreHandleXXXTask`,
`HandleXXXTask` and `PostHandleXXXTask`. The Front Modules consist of `Approver`, `Downloader` and `Uploader`.

### Background Modules

The Background Modules are responsible for handling internal tasks of SP, which are generated internally and thus have
guaranteed information correctness. As a result, there is only one interface `HandleXXXTask` for these tasks. The Background
Modules consist of `Authenticator`, `TaskExecutor`, `Manager`, `P2P`, `Receiver` and `Signer`.

### Module Type

The GfSp framework comprises several modules, including `Gater`, `Authenticator`, `Authorizer`, `Uploader`, `Downloader`,
`Manager`, `P2P`, `TaskExecutor`, `Receiver`, `Signer`, `Metadata` and `BlockSyncer`. Additionally, the GfSp framework
supports the extension of customized modules as required. Once registered in the GfSp framework and executing the
modular interface, these customized modules will be initialized and scheduled.

### Gater

Gater module serves as the gateway for SP, providing HTTP services and adhering to the S3 protocol. It generates tasks
corresponding to user requests and forwards them to other modules within SP. Since Gater does not allow customization,
no interface is defined in the modular file.

### Authenticator

Authenticator module is responsible for verifying authentication.

### Approver

Approver module is responsible for handling approval requests, specifically `CreateBucketApproval` and `CreateObjectApproval`.

### Uploader

Uploader module handles the put object requests from user accounts and stores payload data into piece store of the primary SP.

### Downloader

Downloader module is responsible for handling get object request from user account and get challenge info request from
other components in the Greenfield system.

### TaskExecutor

TaskExecutor module is responsible for handling background task. This module can request tasks from the Manager module,
execute them and report the results or status back to the Manager. The tasks it can handle include ReplicatePieceTask,
SealObjectTask, ReceivePieceTask, GCObjectTask, GCZombiePieceTask, and GCMetaTask.

### Manager

Manager module is responsible for managing task scheduling of SP and other management functions.

### P2P

P2P module is responsible for handling the interaction of control information between SPs. It handles ask replicate piece 
approval requests by broadcasting the approval to other SPs, waiting for responses and returning the approvals if the 
minimum or maximum approved number is reached before the timeout.

### Receiver

Receiver module receives data from the primary SP, calculates the integrity hash of the data, signs it, and returns it
to the primary SP for sealing on a greenfield.

### Signer

Signer module handles the signing of the SP data on the Greenfield chain operator and holds all of the SP's private keys.
Due to the sequence number of the SP account, it must be a singleton.
