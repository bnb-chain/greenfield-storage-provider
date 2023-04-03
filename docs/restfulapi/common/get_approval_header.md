# Get Approval Header Parameter

## MsgCreateBucket

| ParameterName     | Type                           | Description                                                                                                                                                                                                         |
| ----------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Creator           | string                         | Creator is the account address of bucket creator, it is also the bucket owner.                                                                                                                                      |
| BucketName        | string                         | BucketName is a globally unique name of bucket.                                                                                                                                                                     |
| Visibility        | [VisibilityType](#visibilitytype) | visibility means the bucket is private or public. If private, only bucket owner or grantee can read it, otherwise every greenfield user can read it.                                                                |
| PaymentAddress    | string                         | PaymentAddress is an account address specified by bucket owner to pay the read fee. Default: creator.                                                                                                               |
| PrimarySpAddress  | string                         | PrimarySpAddress  is the address of primary sp.                                                                                                                                                                     |
| PrimarySpApproval | [Approval](#approval)             | PrimarySpApproval is the approval info of the primary SP which indicates that primary sp confirm the user's request.                                                                                                |
| ChargedReadQuota  | unsigned integer               | ChargedReadQuota defines the read data that users are charged for, measured in bytes. The available read data for each user is the sum of the free read data provided by SP and the ChargeReadQuota specified here. |

## MsgCreateObject

| ParameterName              | Type                           | Description                                                                                                                                                  |
| -------------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Creator                    | string                         | Creator is the account address of object uploader.                                                                                                           |
| BucketName                 | string                         | BucketName is the name of the bucket where the object is stored.                                                                                             |
| ObjectName                 | string                         | ObjectName is the name of object.                                                                                                                            |
| PayloadSize                | integer                        | PayloadSize is size of the object's payload.                                                                                                                 |
| Visibility                 | [VisibilityType](#visibilitytype) | VisibilityType means the object is private or public. If private, only object owner or grantee can access it, otherwise every greenfield user can access it. |
| ContentType                | string                         | ContentType is a standard MIME type describing the format of the object.                                                                                     |
| PrimarySpApproval          | [Approval](#approval)             | PrimarySpApproval is the approval info of the primary SP which indicates that primary sp confirm the user's request.                                         |
| ExpectChecksums            | byteArray                      | ExpectChecksums is a list of hashes which was generate by redundancy algorithm.                                                                              |
| RedundancyType             | [RedundancyType](#redundancytype) | RedundancyType specifies which redundancy type is used.                                                                                                      |
| ExpectSecondarySpAddresses | stringArray                    | ExpectSecondarySpAddresses is a list of StorageProvider address which is optional.                                                                           |

## Approval

| ParameterName | Type      | Description                               |
| ------------- | --------- | ----------------------------------------- |
| ExpiredHeight | integer   | ExpiredHeight is expired at which height. |
| Sig           | byteArray | Sig is signature                          |

## RedundancyType

| Value | Description                      |
| ----- | -------------------------------- |
| 0     | Redundancy type is replica type. |
| 1     | Redundancy type is ec type.      |

## VisibilityType

| Value | Description                     |
| ----- | ------------------------------- |
| 0     | Visibility type is unspecified. |
| 1     | Visibility type is public read. |
| 2     | Visibility type is private.     |
| 3     | Visibility type is inherit.     |

**Note** If the bucket visibility is inherit, it's finally set to private. If the object Visibility is inherit, it's the same as bucket.
