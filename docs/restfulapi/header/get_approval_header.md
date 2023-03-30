# Get Approval Header Parameter

## MsgCreateBucket

| ParameterName     | Type                  |
| ----------------- | --------------------- |
| Creator           | string                |
| BucketName        | string                |
| IsPublic          | bool                  |
| PaymentAddress    | string                |
| PrimarySpAddress  | string                |
| PrimarySpApproval | [Approval](#approval) |
| ReadQuota         | integer               |

## MsgCreateObject

| ParameterName              | Type                              |
| -------------------------- | --------------------------------- |
| Creator                    | string                            |
| BucketName                 | string                            |
| ObjectName                 | string                            |
| PayloadSize                | integer                           |
| IsPublic                   | bool                              |
| ContentType                | string                            |
| PrimarySpApproval          | [Approval](#approval)             |
| ExpectChecksums            | byteArray                         |
| RedundancyType             | [RedundancyType](#redundancytype) |
| ExpectSecondarySpAddresses | stringArray                       |

## Approval

| ParameterName | Type      |
| ------------- | --------- |
| ExpiredHeight | integer   |
| Sig           | byteArray |

## RedundancyType

| Value | Description  |
| ----- | ------------ |
| 0     | replica type |
| 1     | ec type      |
