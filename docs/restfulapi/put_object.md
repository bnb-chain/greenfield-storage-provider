# PutObject

## RESTful API Description

This API is used to upload an object to Greenfield SP.

## HTTP Request Format

| Desscription | Definition                   |
| ------------ | ---------------------------- |
| Host         | BucketName.gnfd.nodereal.com |
| Path         | /ObjectName                  |
| Method       | PUT                          |
| Accept       | application/xml              |

You should set `BucketName` in url host to upload an object.

`Content-Type` is determined by specific object, such as the content type of an image could be image/jpeg.

## HTTP Request Header

| ParameterName   | Type   | Required | Description                                                             |
| --------------- | ------ | -------- | ----------------------------------------------------------------------- |
| X-Gnfd-Txn-Hash | string | yes      | The transaction hash of the Greenfield chain creates object transaction |
| Authorization   | string | yes      | The authorization string of the HTTP request                            |

## HTTP Request Parameter

### Path Parameter

None

### Query Parameter

None

### Request Body

The request body is a binary data that you want to store in Greenfield SP.

## Request Syntax

```shell
PUT /ObjectName HTTP/1.1
Host: BucketName.gnfd.nodereal.com
X-Gnfd-Txn-Hash: Txn-Hash
Authorization: Authorization

Body
```

## HTTP Response Header

| ParameterName     | Type   | Required | Description                           |
| ----------------- | ------ | -------- | ------------------------------------- |
| X-Gnfd-Request-ID | string | yes      | defines trace id, trace request in sp |
| Etag              | string | yes      | Entity tag for the uploaded object    |

## HTTP Response Parameter

### Response Body

If you failed to send request to put object, you will get reponse body in XML:

| ParameterName | Type   | Description                        |
| ------------- | ------ | ---------------------------------- |
| errorCode     | string | error returned code                |
| errorMessage  | string | the message of error returned code |

## Response Syntax

```shell
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
Etag: Etag
```

## Examples

### Example 1: Upload an object

```shell
PUT /my-image.jpg HTTP/1.1
Host: myBucket.gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
Authorization: authorization string
Content-Type: image/jpeg
Content-Length: 11434
X-Gnfd-Txn-Hash: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
[11434 bytes of object data]
```

### Sample Response: Upload an object successfully

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 4208447844380058399
Date: Fri, 31 March 2023 17:32:10 GMT
ETag: "1b2cf535f27731c974343645a3985328"
Content-Length: 11434
```
