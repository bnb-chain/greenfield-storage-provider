# Core

This directory defines some interfaces, and community developers can customize own 
implementations, replace the default implementations, and realize the customizable 
goals of community developers.

# Concept

## GfSp Framework
GfSp Framework(GfSp) is a base framework for greenfield storage provider. The GfSp
implements the specification of SP. For example, user must ask create object approval 
before uploading object, upload object to primary SP and replicate object to secondary 
SPs are divided into two stages, etc. Under the specification, each process is 
completed through modular, and the modular can be customized. GfSp also provides a 
default implementation, community developers can customize their own implementation 
or optimize the existing implementations. For example, the community can customize 
their own policy of agree approval, upload object and cache the payload piece to speed 
up the performance of replication, and so on.


## Interface
Core directory defines three types interfaces for customization: 
`Infrastructure interface`, `Special Modular` and `Common Modular`.

### Infrastructure interface
Infrastructure interface includes:
* [Consensus](./consensus/consensus.go): is the interface to query greenfield consensus 
  data. the consensus data can come from validator, fullnode, or other off-chain data 
  service
* [ResourManager](./rcmgr/README.md): ResourceManager is the interface to the resource 
  management subsystem. The ResourceManager tracks and accounts for resource usage in 
  the stack, from the internals to the application, and provides a mechanism to limit
  resource usage according to a user configurable policy.
* [PieceStore](./piecestore/piecestore.go): PieceStore is the interface to piece store 
  that store the object payload data.
* [PieceOp](./piecestore/piecestore.go): PieceOp is the helper interface for piece key 
  operator and piece size calculate.
* [SPDB](./spdb/spdb.go): SPDB is the interface to records the SP metadata.
* [BSDB](./bsdb/bsdb.go): BSDB is the interface to records the greenfield chain metadata.
* [TaskQueue](./taskqueue/README.md): Task is the interface to the smallest unit of 
  SP background service interaction. Task scheduling and execution are directly related 
  to the order of task arrival, so task queue is a relatively important basic interface 
  used by all modules inside SP.

### Special Modular
* [Approver](./module/README.md) : Approver is the modular to handle ask approval request, 
  handles CreateBucketApproval and CreateObjectApproval.
* [Authorizer](./module/README.md): Authorizer is the modular to authority verification.
* [Downloader](./module/README.md): Downloader is the modular to handle get object request 
  from user account, and get challenge info request from other components in the system.
* [TaskExecutor](./module/README.md): TaskExecutor is the modular to handle background task, 
  it will ask task from Manager modular, handle the task and report the result or status to
  the manager modular includes: ReplicatePieceTask, SealObjectTask, ReceivePieceTask, 
* GCObjectTask, GCZombiePieceTask, GCMetaTask.
* [Manager](./module/README.md): Manager is the modular to SP's manage modular, it is Responsible 
  for task scheduling and other management of SP.
* [P2P](./module/README.md): P2P is the modular to the interaction of control information 
  between Sps, handles the ask replicate piece approval, it will broadcast the approval to 
  other SPs, wait the responses, if up to min approved number or max approved number before 
  timeout, will return the approvals.
* [Receiver](./module/README.md): Receiver is the modular to receive the piece data from 
  primary SP, calculates the integrity hash of the piece data and sign it, returns to the 
  primary SP for sealing object on greenfield.
* [Signer](./module/README.md): Signer is the modular to handle the SP's sign and on greenfield 
  chain operator. It holds SP all private key. Considering the sp account's sequence number, it
  must be a singleton.
* [Uploader](./module/README.md): Uploader is the modular to handle put object request from user 
  account, and store it in primary SP's piece store.

### Common Modular
In addition to the modular specified above, developers can also customize own modular, 
just implement the [Moduler](./module/modular.go) interface and register it to GfSp.


# Example

## Customized infrastructure interface
```go

// new your own CustomizedPieceStore instance that implement the PieceStore interface
pieceStore := NewCustomizedPieceStore(...)

// new GfSp framework app
gfsp, err := NewGfSpBaseApp(GfSpConfig, CustomizePieceStore(pieceStore))
if err != nil {
    return err
}

gfsp.Start(ctx)

// the GfSp framework will replace the default PieceStore with CustomizedPieceStore
```

## Customized Special Modular
```go
// new your own CustomizedApprover instance that implement the Approver interface
//  NewCustomizedApprover must be func type: 
//      func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error)
approver := NewCustomizedApprover(GfSpBaseApp, GfSpConfig)

// the Special Modular name is Predefined
gfspapp.RegisterModularInfo(model.ApprovalModularName, model.ApprovalModularDescription, approver)

// new GfSp framework app
gfsp, err := NewGfSpBaseApp(GfSpConfig, CustomizeApprover(approver))
if err != nil {
    return err
}

gfsp.Start(ctx)
// the GfSp framework will replace the default Approver with CustomizedApprover
```

## Implement Common Modular

```go

// metadata should implement Modular interface

// register metadata modular to GfSp framework

gfspapp.RegisterModularInfo(MetadataModularName, MetadataModularDescription, NewMetadataModular)


// NewMetadataModular must be func type: 
//  func(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error)

// new GfSp framework app
gfsp, err := NewGfSpBaseApp(GfSpConfig, CustomizeApprover(approver))
if err != nil {
	return err
}

gfsp.Start(ctx)
// the GfSp framework will call the NewMetadataModular and start the metadata

```