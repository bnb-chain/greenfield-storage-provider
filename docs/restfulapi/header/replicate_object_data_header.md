# ReplicateObjectData Header

## ObjectInfo

| ParameterName        | Type                                                      | Description                                                                                             |
| -------------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Owner                | string                                                    | creator is the account address of bucket creator, it is also the bucket owner.                          |
| BucketName           | string                                                    | BucketName is the name of the bucket                                                                    |
| ObjectName           | bool                                                      | ObjectName is the name of object                                                                        |
| Id                   | unsigned integer                                          | Id is the unique identifier of object.                                                                  |
| PayloadSize          | unsigned integer                                          | PayloadSize is the total size of the object payload                                                     |
| IsPublic             | bool                                                      | IsPublic defines the highest permissions for object. When the object is public, everyone can access it. |
| ContentType          | string                                                    | ContentType defines the format of the object which should be a standard MIME type.ReadQuota.            |
| CreateAt             | integer                                                   | CreateAt defines the block number when the object created.                                              |
| ObjectStatus         | [ObjectStatus](#objectstatus)                             | ObjectStatus defines the upload status of the object.                                                   |
| RedundancyType       | [RedundancyType](./get_approval_header.md#redundancytype) | RedundancyType define the type of the redundancy which can be multi-replication or EC.                  |
| SourceType           | [SourceType](#sourcetype)                                 | SourceType defines the source of the object.                                                            |
| Checksums            | byteArray                                                 | Checksums defines the root hash of the pieces which stored in a SP.                                     |
| SecondarySpAddresses | stringArray                                               | SecondarySpAddresses defines the addresses of secondary sps                                             |

## ReplicateApproval

| ParameterName     | Type                      | Description                                              |
| ----------------- | ------------------------- | -------------------------------------------------------- |
| ObjectInfo        | [ObjectInfo](#objectinfo) | ObjectInfo defines the object info for getting approval. |
| SpOperatorAddress | string                    | SpOperatorAddress defines sp operator public key.        |
| ExpiredTime       | integer                   | ExpiredTime defines the approval valid deadline.         |
| RefusedReason     | string                    | RefusedReason defines the reason of refusing.            |
| Signature         | byteArray                 | Signature defines the reason of refusing.                |

## ObjectStatus

| Value | Description           |
| ----- | --------------------- |
| 0     | object status created |
| 1     | object status sealed  |

## SourceType

| Value | Description                    |
| ----- | ------------------------------ |
| 0     | source type is origin          |
| 1     | source type is bsc cross chain |
