# Restful API
This guide explains the Greenfield Storage Provider application programming interface (API). 
It describes various API operations, related request and response structures.
We recommend that you use the [Greenfield Go SDK](https://github.com/bnb-chain/greenfield-go-sdk) 
or the [Greenfield CMD](https://github.com/bnb-chain/greenfield-cmd).

## GetApproval
### Request Syntax
```http request
GET /greenfield/admin/v1/get-approval?action=ActionName
X-Gnfd-Unsigned-Msg: UnsignedMsg
Authorization: AuthorizationMsg
```

### Request Parameters
The request uses the following parameters.
* ActionName

  The action that you want to get approval. 

  Valid Values: `CreateBucket` | `CreateObject`
* UnsignedMsg

  The marshal string of the `CreateBucket` | `CreateObject` Message.
* AuthorizationMsg

  The authorization string of the HTTP request.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
X-Gnfd-Signed-Msg: SignedMsg
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
The response returns the following HTTP headers.
* RequestID
  
  Request ID of the request.
* SignedMsg

  The marshal string of the signed `CreateBucket` | `CreateObject` Message.

## PutObject
### Request Syntax
```http request
PUT /ObjectName HTTP/1.1
Host: BucketName.gnfd.nodereal.com
X-Gnfd-Txn-Hash: Txn-Hash
Authorization: AuthorizationMsg

Body
```

### Request Parameters
The request uses the following parameters.
* ObjectName

  ObjectName for which the PUT action was initiated.
* BucketName

  The bucket name to which the PUT action was initiated.
* Txn-Hash

  The transaction hash of the Greenfield chain create object transaction.
* AuthorizationMsg

  The authorization string of the HTTP request.
* Body

  The request accepts the body binary data.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
ETag: ETag
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
The response returns the following HTTP headers.
* RequestID

  Request ID of the request.
* ETag

  Entity tag for the uploaded object.

## GetObject
### Request Syntax
```http request
Get /ObjectName
Host: BucketName.gnfd.nodereal.com
Authorization: AuthorizationMsg
```

### Request Parameters
The request uses the following parameters.
* ObjectName

  ObjectName for which the PUT action was initiated.
* BucketName

  The bucket name to which the PUT action was initiated.
* AuthorizationMsg

  The authorization string of the HTTP request.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID

Body
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
* RequestID

  Request ID of the request.
* Body

  The response binary data.

## ChallengeObjectData
### Request Syntax
```http request
GET /greenfield/admin/v1/challenge
X-Gnfd-Object-ID: ObjectID
X-Gnfd-Redundancy-Index: RedundancyIndex
X-Gnfd-Piece-Index: PieceIndex
Authorization: AuthorizationMsg
```

### Request Parameters
The request uses the following parameters.
* ObjectID

  Object ID of the challenged object.
* RedundancyIndex

  Redundancy Index of the challenged object which is used to specify the SP.

* PieceIndex

  Piece Index is used to specify the object's piece data.
* AuthorizationMsg

  The authorization string of the HTTP request.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
X-Gnfd-Integrity-Hash: IntegrityHash
X-Gnfd-Piece-Hash: PieceHashList

Body
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
* RequestID

  Request ID of the request.
* IntegrityHash

  Integrity Hash which the SP recorded.
* PieceHashList

  Piece Hash List which the SP recorded.
* Body

  The piece binary data.

## QueryBucketReadQuota
### Request Syntax
```http request
GET /?read-quota&year-month=YearMonth
Host: BucketName.gnfd.nodereal.com
Authorization: AuthorizationMsg
```

### Request Parameters
The request uses the following parameters.
* YearMonth

  YearMonth is used to specify queried month, format "2023-03".
* BucketName

  The bucket name is used to specify queried bucket.
* AuthorizationMsg

  The authorization string of the HTTP request.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
Content-Type: application/xml

Body
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
* RequestID

  Request ID of the request.
* Body
  ```xml
  <GetBucketReadQuotaResult>
    <BucketName>name</BucketName>
    <BucketID>id</BucketID>
    <ReadQuotaSize>quota_size</ReadQuotaSize>
    <SPFreeReadQuotaSize>sp_free_quota_size</SPFreeReadQuotaSize>
    <ReadConsumedSize>consumed_size</ReadConsumedSize>
  </GetBucketReadQuotaResult>
  ```
  
## ListBucketReadRecords
### Request Syntax
```http request
GET /?list-read-record&max-records=MaxRecord&start-timstamp=StartTimestamp&end-timestamp=End-Timestqamp
Host: BucketName.gnfd.nodereal.com
Authorization: AuthorizationMsg
```

### Request Parameters
The request uses the following parameters.
* MaxRecord

  MaxRecord is used to specify the max list number.
* StartTimestamp

  StartTimestamp is used to specify start microsecond timestamp, which the time range is [start, end).
* EndTimestamp

  EndTimestamp is used to specify end microsecond timestamp, which the time range is [start, end).
* BucketName

  The bucket name is used to specify queried bucket.
* AuthorizationMsg

  The authorization string of the HTTP request.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
Content-Type: application/xml

Body
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
* RequestID

  Request ID of the request.
* Body
  ```xml
  <ListBucketReadRecordResult>
    <NextStartTimestampUs>ts</NextStartTimestampUs>
    <ReadRecords>
      <ObjectName>name</ObjectName>
      <ObjectID>id</ObjectID>
      <ReadAccountAddress>address</ReadAccountAddress>
      <ReadTimestampUs>timestamp</ReadTimestampUs>
      <ReadSize>size</ReadSize>
    </ReadRecords>
    ...
  </ListBucketReadRecordResult>
  ```

## ReplicateObjectData
### Request Syntax
```http request
PUT /greenfield/receiver/v1/sync-piece
X-Gnfd-Object-Info: ObjectInfo
X-Gnfd-Replica-Idx: ReplicaIdx
X-Gnfd-Segment-Size: SegmentSize
X-Gnfd-Replica-Approval: ReplicaApproval

Body
```

### Request Parameters
The request uses the following parameters.
* ObjectInfo

  The marshal string of the `ObjectInfo` Message.
* ReplicaIdx

  The index of SP, which will be changed to `RedundancyIndex`.
* SegmentSize

  The piece size of replicated data, which will be changed to `PieceSize`.
* ReplicaApproval

  The replicated approval, which will be changed to `ReplicateApproval`.
* Body

  The replicated binary data.

### Response Syntax
```http request
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
X-Gnfd-Integrity-Hash: IntegrityHash
X-Gnfd-Integrity-Hash-Signature: IntegrityHashSignature
```

### Response Elements
If the request is successful, the service sends back an HTTP 200 response.
* RequestID

  Request ID of the request.
* IntegrityHash

  The integrity hash of the replicated data.
* IntegrityHashSignature

  The integrity hash signature of the replicated data.