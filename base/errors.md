# GfSpErrors

GfSpErrors consists of four parts: CodeSpace, HttpStatusCode, InnerCode and 
Description, where HttpStatusCode and InnerCode combined Description as error
message are returned to the user.

## InnerCode 
InnerCode is an error code in the SP system to help developers quickly locate 
problems. It is assembled from three parts, and each part occupying two digits.
The first two represent the module code of the error occurs, the middle two 
digits represent the dependent module code reports an error, the last two digits
represent local error code.

Example: `105001` stands for the `10` modular calls `50` modular and the `50` 
modular reports an error, and in the `10` modular it is the `01` error.

## Reserved module codes

### Infrastructure Code
It sorts in reverse order starting from 99...
* `99`: is used for GfSp Bass App code.
* `98`: is used for GfSp Client code.

### System External Dependencies Code
It sorts from 50...
* `50`: is used for Consensus code.
* `51`: is used for PieceStore code.
* `52`: is used for SpDB code.
* `53`: is used for Resource manager code.

### Modular Code
It sorts from 0... to 49
* `01`: is used for Approver modular code.
* `02`: is used for Authorizer modular code.
* `03`: is used for Downloader modular code.
* `04`: is used for TaskExecutor modular code.
* `05`: is used for Gateway modular code.
* `06`: is used for Manager modular code.
* `07`: is used for P2P modular code.
* `08`: is used for Receiver modular code.
* `09`: is used for Metadata modular code.
* `10`: is used for Signer modular code.
* `11`: is used for Uploader modular code.

### InnerCode Example
`105001`: stands for Signer calls Consensus occurs an error, in the Signer the local 
error is `01`.
`11002`: stands for Approver calls Signer occurs an error, in the Signer the local
error is `02`.
**notice**: if omit the first 0.

# GfSpErrorsExample
```go
// ErrExceedQueue stands the error of the task queue is exceed.
// InnerCode 10004 stands 01 - Approver, 00-no dependencies other, 04-Approver local the 4th error
var (
    ErrDanglingPointer    = gfsperrors.Register(ApprovalModularName, http.StatusNotFound, 10001, "OoooH.... request lost")
    ErrExceedBucketNumber = gfsperrors.Register(ApprovalModularName, http.StatusServiceUnavailable, 10002, "account buckets exceed the limit")
    ErrRepeatedTask       = gfsperrors.Register(ApprovalModularName, http.StatusBadRequest, 10003, "ask approval request repeated")
    ErrExceedQueue        = gfsperrors.Register(ApprovalModularName, http.StatusServiceUnavailable, 10004, "ask approval request exceed the limit, try again later")
)
```
