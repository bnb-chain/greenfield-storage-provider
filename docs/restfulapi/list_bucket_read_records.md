# ListBucketReadRecords

## RESTful API Description

This API is used to list bucket read records.

## HTTP Request Format

| Desscription | Definition                   |
| ------------ | ---------------------------- |
| Path         | /                            |
| Host         | BucketName.gnfd.nodereal.com |
| Method       | GET                          |
| Accept       | application/xml              |

You should set `BucketName` in url host to determine which bucket do you want to query.

## HTTP Request Header

| ParameterName | Type   | Required | Description                                  |
| ------------- | ------ | -------- | -------------------------------------------- |
| Authorization | string | yes      | The authorization string of the HTTP request |

## HTTP Request Parameter

### Path Parameter

The request does not have a path parameter.

### Query Parameter

| ParameterName    | Type    | Required | Description                                                                                        |
| ---------------- | ------- | -------- | -------------------------------------------------------------------------------------------------- |
| list-read-record | string  | yes      | List read quota path                                                                               |
| max-records      | integer | yes      | max-records is used to specify the max list number, the biggest number is 1000                     |
| start-timstamp   | integer | yes      | start-timstamp is used to specify start microsecond timestamp which the time range is [start, end) |
| end-timestamp    | integer | yes      | EndTimestamp is used to specify end microsecond timestamp which the time range is [start, end)     |

### Request Body

The request does not have a request body.

## Request Syntax

```shell
GET /?list-read-record&max-records=MaxRecord&start-timstamp=StartTimestamp&end-timestamp=End-Timestqamp
Host: BucketName.gnfd.nodereal.com
Authorization: Authorization
```

## HTTP Response Header

| ParameterName     | Type   | Description                           |
| ----------------- | ------ | ------------------------------------- |
| X-Gnfd-Request-ID | string | defines trace id, trace request in sp |
| Content-Type      | string | value is `application/xml`            |

## HTTP Response Parameter

### Response Body

If the request is successful, the service sends back an HTTP 200 response.

The following data is returned in XML format by the service.

| ParameterName      | Type             | Description                                    |
| ------------------ | ---------------- | ---------------------------------------------- |
| ObjectName         | string           | ObjectName is the read object name             |
| ObjectID           | string           | ObjectID is the read object id                 |
| ReadAccountAddress | string           | ReadAccountAddress is the read account address |
| ReadTimestampUs    | integer          | ReadTimestampUs is the read time stamp         |
| ReadSize           | unsigned integer | ReadSize is the read object size               |

If you failed to send request to put object, you will get reponse body in XML:

| ParameterName | Type   | Description                        |
| ------------- | ------ | ---------------------------------- |
| errorCode     | string | error returned code                |
| errorMessage  | string | the message of error returned code |

## Response Syntax

```shell
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID

XML Body
```

## Examples

### Example 1: List bucket read records

```shell
GET /?list-read-record&max-records=2&start-timstamp=1680520105786&end-timestamp=1680520178958 HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
Authorization: authorization string
```

### Sample Response: List bucket read records successfully

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 4208447844380058399
Date: Fri, 31 March 2023 17:32:10 GMT

<?xml version="1.0" encoding="UTF-8"?>
<ListBucketReadRecordResult>
    <NextStartTimestampUs>ts</NextStartTimestampUs>
    <ReadRecords>
      <ObjectName>myObject</ObjectName>
      <ObjectID>836329</ObjectID>
      <ReadAccountAddress>0xewh23289y23hd</ReadAccountAddress>
      <ReadTimestampUs>1680520178958</ReadTimestampUs>
      <ReadSize>10</ReadSize>
    </ReadRecords>
    ...
</ListBucketReadRecordResult>
```
