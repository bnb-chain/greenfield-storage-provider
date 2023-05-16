# Modular

Modular is a complete logical module of SP. The GfSp framework is responsible 
for the necessary interaction between modules. As for the implementation 
of the module, it can be customized. Example, The GfSp framework stipulates 
that ask object approval must be carried out before uploading an object, 
whether agrees the approval, SP can be customized.

# Concept

## Front Modular
Front Modular handles the user's request, the gater will generate corresponding 
task and send to Front Modular, the Front Modular need check the request 
is correct. and after handle the task maybe some extra work is required. 
So the Front Modular has three interfaces for each task type, `PreHandleXXXTask`, 
`HandleXXXTask` and`PostHandleXXXTask`. Front Modular includes: `Approver`, 
`Downloader` and `Uploader`. 

## Background Modular
Background Modular handles the SP inner task, since it is internally 
generated, the correctness of the information can be guaranteed, so only 
have one interface`HandleXXXTask`. Background Modular includes: `Authorizer`,
`TaskExecutor`,`Manager`, `P2P`, `c` and `Signer`.


# Modular Type

The GfSp framework specifies the following modular: `Gater`, `Approver`, 
`Authorizer`, `Uploader`, `Downloader`, `Manager`, `P2P`, `Receiver`, 
`Signer`and `Retriever`. The GfSp framework also supports extending more 
customized mudolar as needed. As long as it is registered in GfSp framework 
and executes the modular interface, it will be initialized and scheduled.

## Gater
Gater as SP's gateway, provides http service and follows the s3 protocol, 
and generates corresponding task and forwards them to other modular in the 
SP. It does not allow customization, so no interface is defined in the 
modular file.

## Approver
Approver is the modular to handle ask approval request, handles CreateBucketApproval 
and CreateObjectApproval.

## Authorizer
Authorizer is the modular to authority verification.

## Downloader
Downloader is the modular to handle get object request from user account,
and get challenge info request from other components in the system.

## TaskExecutor
TaskExecutor is the modular to handle background task, it will ask task 
from Manager modular, handle the task and report the result or status to 
the manager modular includes: ReplicatePieceTask, SealObjectTask, 
ReceivePieceTask, GCObjectTask, GCZombiePieceTask, GCMetaTask.

## Manager
Manager is the modular to SP's manage modular, it is Responsible for task 
scheduling and other management of SP.

## P2P
P2P is the modular to the interaction of control information between Sps, 
handles the ask replicate piece approval, it will broadcast the approval 
to other SPs, wait the responses, if up to min approved number or max 
approved number before timeout, will return the approvals.

## Receiver
Receiver is the modular to receive the piece data from primary SP, calculates
the integrity hash of the piece data and sign it, returns to the primary SP 
for sealing object on greenfield.

## Signer
Signer is the modular to handle the SP's sign and on greenfield chain operator. 
It holds SP all private key. Considering the sp account's sequence number, it 
must be a singleton.

## Uploader
Uploader is the modular to handle put object request from user account, and
store it in primary SP's piece store. 