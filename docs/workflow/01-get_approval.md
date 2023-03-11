# Get Approval

Before CreateBucket, PutObject, replicate Object to the secondary SPs, the request initiator need send a GetApproval request to 
ask whether the SP is willing to serve the related objects. The SP can consider whether it is willing to accept approval based 
on some dimensions, such ad bucket, object and user, eg: SP can reject users with bad reputation, and can reject specific objects 
or buckets. The SP acknowledges the request by signing a message about the operation and response to the initiator, if the SP does 
not want to serve(the default policy is to serve, each SP can customize its own strategy), it can refuse to sign.

## Gateway
* Receives the GetApproval request from the request initiator.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account is existed.
* Fills the CreateBucket/PutObject/ReplicateObjectData message's timeout field and dispatches the request to Signer service.
* Gets Signature from Signer and fill the message's approval signature field, and returns to the request initiator.

### GetApproval to the primary SP
```protobuf
message Approval {
  uint64 expired_height = 1;
  bytes sig = 2;
}
message MsgCreateBucket {
  option (cosmos.msg.v1.signer) = "creator";
  // creator is the account address of bucket creator, it is also the bucket owner.
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is a globally unique name of bucket
  string bucket_name = 2;
  // is_public means the bucket is private or public. if private, only bucket owner or grantee can read it,
  // otherwise every greenfield user can read it.
  bool is_public = 3;
  // payment_address is an account address specified by bucket owner to pay the read fee. Default: creator
  string payment_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_address is the address of primary sp.
  string primary_sp_address = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_approval is the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval primary_sp_approval = 7;
  // read_quota
  uint64 read_quota = 8;
}
message MsgCreateObject {
  option (cosmos.msg.v1.signer) = "creator";
  // creator is the account address of object uploader
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name is the name of object
  string object_name = 3;
  // payload_size is size of the object's payload
  uint64 payload_size = 4;
  // is_public means the bucket is private or public. if private, only bucket owner or grantee can access it,
  // otherwise every greenfield user can access it.
  bool is_public = 5;
  // content_type is a standard MIME type describing the format of the object.
  string content_type = 6;
  // primary_sp_approval is the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval primary_sp_approval = 7;
  // expect_checksums is a list of hashes which was generate by redundancy algorithm.
  repeated bytes expect_checksums = 8;
  // redundancy_type can be ec or replica
  RedundancyType redundancy_type = 9;
  // expect_secondarySPs is a list of StorageProvider address, which is optional
  repeated string expect_secondary_sp_addresses = 10 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

```


### GetApproval to the secondary SP
```protobuf

message ReplicateApproval {
  uint64 expired_time = 1;
  bytes sig = 2;
}
message MsgReplicateObjectData {
  // object_info defines the object info for getting approval
  bnbchain.greenfield.storage.ObjectInfo object_info = 1;
  // replicate_approval is the approval info of replicating object
  ReplicateApproval replicate_approval = 2;
}

```
* The SP receives the CreateBucket/CreateObject/ReplicateObjectData GetApproval request.
  * If refuse to serve(the default policy is to serve, each SP can customize its own strategy), the SP returns a refused response and reason.
  * If willing to serve, the SP add the expired-height/expired-time  field to the Approval field and sign it, and returns.

## Signer
* Receives the CreateBucket/PutObject/ReplicateObjectData message, sign it by using the approval private key and response to the Gateway service.
