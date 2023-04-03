# GetApproval

## RESTful API Description

This API is used to sign an approval for creating a bucket action or creating an object action.

## HTTP Request Format

| Desscription | Definition                        |
| ------------ | --------------------------------- |
| Path         | /greenfield/admin/v1/get-approval |
| Method       | GET                               |
| Content-Type | application/xml                   |
| Accept       | application/xml                   |

## HTTP Request Header

| ParameterName                                       | Type   | Required | Description                                  |
| --------------------------------------------------- | ------ | -------- | -------------------------------------------- |
| [X-Gnfd-Unsigned-Msg](./header/get_approval_header.md) | string | yes      | defines unsigned msg                         |
| Authorization                                       | string | yes      | The authorization string of the HTTP request |

## HTTP Request Parameter

### Path Parameter

The request does not have a path parameter.

### Query Parameter

| ParameterName | Type   | Required | Description                                                 |
| ------------- | ------ | -------- | ----------------------------------------------------------- |
| action        | string | yes      | The action of approval:`CreateBucket` or `CreateObject` |

### Request Body

The request does not have a request body.

## Request Syntax

```shell
GET /greenfield/admin/v1/get-approval?action=action HTTP/1.1
Content-Type: ContentType
X-Gnfd-Unsigned-Msg: UnsignedMsg
Authorization: Authorization
```

## HTTP Response Header

The response returns the following HTTP headers.

| ParameterName                                     | Type   | Description                           |
| ------------------------------------------------- | ------ | ------------------------------------- |
| X-Gnfd-Request-ID                                 | string | defines trace id, trace request in sp |
| [X-Gnfd-Signed-Msg](./header/get_approval_header.md) | string | defines signed msg                    |

## HTTP Response Parameter

### Response Body

If the request is successful, the service sends back an HTTP 200 response.

If you failed to send request to get approval, you will get error response body in [XML](./common/error.md#sp-error-response-parameter).

## Response Syntax

```shell
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
X-Gnfd-Signed-Msg: SignedMsg
```

## Examples

### Example 1: Create bucket

The following request sends `CreateBucket` action to get approval.

```shell
GET /greenfield/admin/v1/get-approval?action=CreateBucket HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
X-Gnfd-Unsigned-Msg: unsigned msg string
Authorization: authorization string
```

### Sample Response: Create bucket successfully

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 14779951378820359452
X-Gnfd-Signed-Msg: df5857b2ac67b491ba6d9c6632618be7fb22de13662356b593d74103408cf1af46eed90edaa77bdb65b12fc63ee3bec8314ad7bb0f3ae099ccf7dafe22abff2e01
```

## Example 2: Create object

The following request sends `CreateObject` action to get approval.

```shell
GET /greenfield/admin/v1/get-approval?action=CreateObject HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
X-Gnfd-Unsigned-Msg: unsigned msg string
Authorization: authorization string
```

### Sample Response: Create object successfully

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 4208447844380058399
X-Gnfd-Signed-Msg: f00daace3251076f270984e596bbd72b1b1f2a1ae0443e6f32f37cef73d541d568a542333f6a9af2f235724d2a763b3cdc0b370d978d0315b8414fa51fc32a2e00
```

## Example 3: Failed to create bucket

The following request sends `CreateBucket` action to get approval.

```shell
GET /greenfield/admin/v1/get-approval?action=CreateBucket HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
X-Gnfd-Unsigned-Msg: unsigned msg string
Authorization: authorization string
```

## Sample Response: There is an internal error in SP server

```shell
HTTP/1.1 403 Forbidden

<Error>
    <Code>InvalidUnsignedMsg</Code>
    <Message>The uinsigned message is not valid for creating bucket</Message>
    <RequestId>14379357152578345503</RequestId>
</Error>
```
