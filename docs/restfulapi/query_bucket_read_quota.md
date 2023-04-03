# QueryBucketReadQuota

## RESTful API Description

This API is used to query bucket read quota.

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

None

### Query Parameter

| ParameterName | Type   | Required | Description                                                  |
| ------------- | ------ | -------- | ------------------------------------------------------------ |
| read-quota    | string | yes      | Read quota path                                              |
| year-month    | string | yes      | YearMonth is used to specify queried month, format "2023-03" |

### Request Body

None

## Request Syntax

```shell
GET /?read-quota&year-month=YearMonth
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

| ParameterName       | Type    | Description                                                         |
| ------------------- | ------- | ------------------------------------------------------------------- |
| BucketName          | string  | bucket name                                                         |
| BucketID            | string  | bucket id                                                           |
| ReadQuotaSize       | integer | ReadQuotaSize is the greenfield chain bucket info's read quota size |
| SPFreeReadQuotaSize | integer | SPFreeReadQuotaSize is the sp default free quota                    |
| ReadConsumedSize    | integer | ReadConsumedSize is currently consumed size                         |

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

### Example 1: Query a bucket read quota

```shell
GET /?read-quota&year-month=2023-03 HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
Authorization: authorization string
```

### Sample Response: Query a bucket read quota successfully

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 4208447844380058399
Date: Fri, 31 March 2023 17:32:10 GMT

<?xml version="1.0" encoding="UTF-8"?>
<GetBucketReadQuotaResult>
    <BucketName>myBucket</BucketName>
    <BucketID>6u754</BucketID>
    <ReadQuotaSize>20</ReadQuotaSize>
    <SPFreeReadQuotaSize>10</SPFreeReadQuotaSize>
    <ReadConsumedSize>5</ReadConsumedSize>
</GetBucketReadQuotaResult>
```
