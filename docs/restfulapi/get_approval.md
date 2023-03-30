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

| ParameterName                                          | Type   | Required | Description                                                     |
| ------------------------------------------------------ | ------ | -------- | --------------------------------------------------------------- |
| [X-Gnfd-Unsigned-Msg](./header/get_approval_header.md) | string | YES      | defines unsigned msg                                            |
| Authorization                                          | string | YES      | defines authorization, verify signature and check authorization |

## HTTP Request Parameter

### Path Parameter

None

### Query Parameter

| ParameterName | Type   | Required | Description                                              |
| ------------- | ------ | -------- | -------------------------------------------------------- |
| action        | string | YES      | The action of approval: `CreateBucket` or `CreateObject` |

### Request Body

None

## Request Syntax

```shell
GET /greenfield/admin/v1/get-approval?action=action HTTP/1.1
Content-Type: ContentType
X-Gnfd-Unsigned-Msg: UnsignedMsg
Authorization: Authorization
```

## HTTP Response Header

| ParameterName                                        | Type   | Required | Description                           |
| ---------------------------------------------------- | ------ | -------- | ------------------------------------- |
| X-Gnfd-Request-ID                                    | string | YES      | defines trace id, trace request in sp |
| [X-Gnfd-Signed-Msg](./header/get_approval_header.md) | string | YES      | defines signed msg                    |

## HTTP Response Parameter

### Reponse Body

If you failed to send request to get-approval, you will get reponse body in XML:


| ParameterName | Type   | Description                        |
| ------------- | ------ | ---------------------------------- |
| errorCode     | string | error returned code                |
| errorMessage  | string | the message of error returned code |

## Response Syntax

```shell
HTTP/1.1 200
X-Gnfd-Request-ID: RequestID
X-Gnfd-Signed-Msg: SignedMsg
```

## Sample Request: CreateBucket

The following request sends `CreateBucket` action to get approval.

```shell
GET /greenfield/admin/v1/get-approval?action=CreateBucket HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
X-Gnfd-Unsigned-Msg: unsigned msg string
Authorization: authorization string
```

## Success Response: CreateBucket

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 14779951378820359452
X-Gnfd-Signed-Msg: df5857b2ac67b491ba6d9c6632618be7fb22de13662356b593d74103408cf1af46eed90edaa77bdb65b12fc63ee3bec8314ad7bb0f3ae099ccf7dafe22abff2e01
```

## Sample Request: CreateObject

The following request sends `CreateObject` action to get approval.

```shell
GET /greenfield/admin/v1/get-approval?action=CreateObject HTTP/1.1
Host: gnfd.nodereal.com
Date: Fri, 31 March 2023 17:32:00 GMT
X-Gnfd-Unsigned-Msg: unsigned msg string
Authorization: authorization string
```

## Success Response: CreateObject

```shell
HTTP/1.1 200 OK
X-Gnfd-Request-ID: 4208447844380058399
X-Gnfd-Signed-Msg: f00daace3251076f270984e596bbd72b1b1f2a1ae0443e6f32f37cef73d541d568a542333f6a9af2f235724d2a763b3cdc0b370d978d0315b8414fa51fc32a2e00
```
