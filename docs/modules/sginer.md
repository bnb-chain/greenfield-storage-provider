# Signer

Signer uses the SP's private keys to sign the message, the messages to form a transaction and sign the transaction to broadcast it to Greenfield BlockChain, or the messages exchanged between SPs.

Signer is an abstract interface to handle the signature of SP and on greenfield chain operator. It holds all private keys of one SP. Considering the SP account's sequence number, it must be a singleton.

```go
type Signer interface {
    Modular
    // SignCreateBucketApproval signs the MsgCreateBucket for asking create bucket approval.
    SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) ([]byte, error)
    // SignCreateObjectApproval signs the MsgCreateObject for asking create object approval.
    SignCreateObjectApproval(ctx context.Context, task *storagetypes.MsgCreateObject) ([]byte, error)
    // SignReplicatePieceApproval signs the ApprovalReplicatePieceTask for asking replicate pieces to secondary SPs.
    SignReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask) ([]byte, error)
    // SignReceivePieceTask signs the ReceivePieceTask for replicating pieces data between SPs.
    SignReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error)
    // SignSecondaryBls signs the secondary bls of object for sealing object.
    SignSecondaryBls(ctx context.Context, objectID uint64, hash [][]byte) ([]byte, error)
    // SignP2PPingMsg signs the ping msg for p2p node probing.
    SignP2PPingMsg(ctx context.Context, ping *gfspp2p.GfSpPing) ([]byte, error)
    // SignP2PPongMsg signs the pong msg for p2p to response ping msg.
    SignP2PPongMsg(ctx context.Context, pong *gfspp2p.GfSpPong) ([]byte, error)
    // SealObject signs the MsgSealObject and broadcast the tx to greenfield.
    SealObject(ctx context.Context, object *storagetypes.MsgSealObject) error
    // RejectUnSealObject signs the MsgRejectSealObject and broadcast the tx to greenfield.
    RejectUnSealObject(ctx context.Context, object *storagetypes.MsgRejectSealObject) error
    // DiscontinueBucket signs the MsgDiscontinueBucket and broadcast the tx to greenfield.
    DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) error
}
```

Signer interface inherits [Modular interface](./common/lifecycle_modular.md#modular-interface), so Uploader module can be managed by lifycycle and resource manager.

In terms of the functions provided by Signer module, there are ten methods. You can rewrite these methods to meet your own requirements.

## SignCreateBucketApproval

The corresponding `protobuf` definition is shown below:

- [MsgCreateBucket](./common/proto.md#msgcreatebucket-proto)

## SignCreateObjectApproval

The corresponding `protobuf` definition is shown below:

- [MsgCreateObject](./common/proto.md#msgcreateobject-proto)

## SignReplicatePieceApproval

The second params of SignReplicatePieceApproval is a task interface, the corresponding interface definition is shown below:

- [ApprovalReplicatePieceTask](./common/task.md#approvalreplicatepiecetask)

The corresponding `protobuf` definition is shown below:

- [GfSpReplicatePieceApprovalTask](./common/proto.md#gfspreplicatepieceapprovaltask-proto)

## SignReceivePieceTask

The second params of SignReceivePieceTask is a task interface, the corresponding interface definition is shown below:

- [ReceivePieceTask](./common/task.md#approvalreplicatepiecetask)

The corresponding `protobuf` definition is shown below:

- [GfSpReceivePieceTask](./common/proto.md#gfspreceivepiecetask-proto)

## SignP2PPingMsg

The corresponding `protobuf` definition is shown below:

- [GfSpPing](./common/proto.md#gfspping-proto)

## SignP2PPongMsg

The corresponding `protobuf` definition is shown below:

- [GfSpPong](./common/proto.md#gfsppong-proto)

## SealObject

The corresponding `protobuf` definition is shown below:

- [MsgSealObject](./common/proto.md#msgsealobject)

## RejectUnSealObject

The corresponding `protobuf` definition is shown below:

- [MsgRejectSealObject](./common/proto.md#msgrejectsealobject-proto)

## DiscontinueBucket

The corresponding `protobuf` definition is shown below:

- [MsgDiscontinueBucket](./common/proto.md#msgdiscontinuebucket)

## GfSp Framework Signer Code

Signer module code implementation: [Signer](https://github.com/bnb-chain/greenfield-storage-provider/tree/master/modular/signer)
