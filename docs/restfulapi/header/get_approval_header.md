# Get Approval Header Parameter

## MsgCreateBucket

| ParameterName     | Type                  | Description                                                                                                                                        |
| ----------------- | --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Creator           | string                | creator is the account address of bucket creator, it is also the bucket owner.                                                                     |
| BucketName        | string                | BucketName is a globally unique name of bucket                                                                                                     |
| IsPublic          | bool                  | IsPublic means the bucket is private or public. if private, only bucket owner or grantee can read it, otherwise every greenfield user can read it. |
| PaymentAddress    | string                | PaymentAddress is an account address specified by bucket owner to pay the read fee. Default: creator                                               |
| PrimarySpAddress  | string                | PrimarySpAddress  is the address of primary sp.                                                                                                    |
| PrimarySpApproval | [Approval](#approval) | PrimarySpApproval is the approval info of the primary SP which indicates that primary sp confirm the user's request.                               |
| ReadQuota         | integer               | ReadQuota                                                                                                                                          |

## MsgCreateObject

| ParameterName              | Type                              | Description                                                                                                                                            |
| -------------------------- | --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Creator                    | string                            | creator is the account address of object uploader                                                                                                      |
| BucketName                 | string                            | BucketName is the name of the bucket where the object is stored.                                                                                       |
| ObjectName                 | string                            | ObjectName is the name of object                                                                                                                       |
| PayloadSize                | integer                           | PayloadSize is size of the object's payload                                                                                                            |
| IsPublic                   | bool                              | IsPublic means the bucket is private or public. if private, only bucket owner or grantee can access it, otherwise every greenfield user can access it. |
| ContentType                | string                            | ContentType is a standard MIME type describing the format of the object.                                                                               |
| PrimarySpApproval          | [Approval](#approval)             | PrimarySpApproval is the approval info of the primary SP which indicates that primary sp confirm the user's request.                                   |
| ExpectChecksums            | byteArray                         | ExpectChecksums is a list of hashes which was generate by redundancy algorithm.                                                                        |
| RedundancyType             | [RedundancyType](#redundancytype) | RedundancyType specifies which redundancy type is used                                                                                                 |
| ExpectSecondarySpAddresses | stringArray                       | ExpectSecondarySpAddresses is a list of StorageProvider address which is optional                                                                      |

## Approval

| ParameterName | Type      | Description                              |
| ------------- | --------- | ---------------------------------------- |
| ExpiredHeight | integer   | ExpiredHeight is expired at which height |
| Sig           | byteArray | Sig is signature                         |

## RedundancyType

| Value | Description  |
| ----- | ------------ |
| 0     | replica type |
| 1     | ec type      |
