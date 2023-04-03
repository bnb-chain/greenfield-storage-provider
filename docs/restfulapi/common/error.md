# SP Error Response Parameter

| ParameterName | Type   | Description   |
| ------------- | ------ | ------------- |
| Code          | string | error code    |
| Message       | string | error message |
| RequestId     | string | request id    |

## Error Response Sample

```xml
<Error>
    <Code>InternalError</Code>
    <Message>rpc error: code = Unavailable desc = connection error: desc = xxx</Message>
    <RequestId>14379357152578345503</RequestId>
</Error>
```
