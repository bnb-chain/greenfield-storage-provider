# ReplicateObjectData

## RESTful API Description

This API is used by primary SP to replicate object data to secondary SP.

## HTTP Request Format

| Desscription | Definition                         |
| ------------ | ---------------------------------- |
| Path         | /greenfield/receiver/v1/sync-piece |
| Host         | gnfd.nodereal.com                  |
| Method       | PUT                                |
| Content-Type | application/octet-stream           |
| Accept       | application/xml                    |

## HTTP Request Header

| ParameterName           | Type   | Required | Description                                                                                                                      |
| ----------------------- | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| X-Gnfd-Object-Info      | string | yes      | The marshal string of the [ObjectInfo](./header/replicate_object_data_header.md#objectinfo) Message                              |
| X-Gnfd-Replica-Idx      | string | yes      | The index of SP which will be changed to `RedundancyIndex`                                                                       |
| X-Gnfd-Segment-Size     | string | yes      | The piece size of replicated data which will be changed to `PieceSize`                                                           |
| X-Gnfd-Replica-Approval | string | yes      | The replicated approval which will be changed to [ReplicateApproval](./header/replicate_object_data_header.md#replicateapproval) |

## HTTP Request Parameter

### Path Parameter

The request does not have a path parameter.

### Query Parameter

The request does not have a query parameter.

### Request Body

The replicated binary data.

## Request Syntax

```shell
PUT /greenfield/receiver/v1/sync-piece
Host: gnfd.nodereal.com
X-Gnfd-Object-Info: ObjectInfo
X-Gnfd-Replica-Idx: ReplicaIdx
X-Gnfd-Segment-Size: SegmentSize
X-Gnfd-Replica-Approval: ReplicaApproval

Body
```

## HTTP Response Header

| ParameterName                   | Type   | Description                                           |
| ------------------------------- | ------ | ----------------------------------------------------- |
| X-Gnfd-Request-ID               | string | defines trace id, trace request in sp                 |
| X-Gnfd-Integrity-Hash           | string | The integrity hash of the replicated data             |
| X-Gnfd-Integrity-Hash-Signature | string | The integrity hash's signature of the replicated data |

## HTTP Response Parameter

### Response Body

If the request is successful, the service sends back an HTTP 200 response.

If you failed to send request to put object, you will get reponse body in XML:

| ParameterName | Type   | Description                        |
| ------------- | ------ | ---------------------------------- |
| errorCode     | string | error returned code                |
| errorMessage  | string | the message of error returned code |

## Response Syntax

```shell
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
X-Gnfd-Integrity-Hash: IntegrityHash
X-Gnfd-Integrity-Hash-Signature: IntegrityHashSignature
```

## Examples

### Example 1: Replica object data

```shell
GET /greenfield/receiver/v1/sync-piece HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
X-Gnfd-Object-Info: ObjectInfo
X-Gnfd-Replica-Idx: ReplicaIdx
X-Gnfd-Segment-Size: SegmentSize
X-Gnfd-Replica-Approval: ReplicaApproval

[14194304 bytes of object data]
```

### Sample Response: List bucket read records successfully

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 4208447844380058399
X-Gnfd-Integrity-Hash: b60a9f213e55e99e8d010b1eb76929c294097aefa623ec1dffe2f67035df0726
X-Gnfd-Integrity-Hash-Signature: IntegrityHash89d1b5abefad08a67e76ef99aad4402cb2e01874936b834207a57e7215e2d4352de95922c2e2542d78141d278787e1163d42c13a43637f2f21f786e767a41dcb01Signature
```
