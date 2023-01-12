package gateway

const (
	// path
	AdminPath          = "/greenfield/admin/v1/"
	GetApprovalSubPath = "get-approval"

	// query key
	TransactionQuery = "transaction"
	PutObjectV2Query = "putobjectv2"
	ActionQuery      = "action"

	// bfs header key
	BFSRequestIDHeader       = "X-Bfs-Request-ID"
	BFSContentLengthHeader   = "X-Bfs-Content-Length"
	BFSContentTypeHeader     = "X-Bfs-Content-Type"
	BFSChecksumHeader        = "X-Bfs-Checksum"
	BFSIsPrivateHeader       = "X-Bfs-Is-Private"
	BFSTransactionHashHeader = "X-Bfs-Txn-Hash"
	BFSResourceHeader        = "X-Bfs-Resource"
	BFSPreSignatureHeader    = "X-Bfs-Pre-Signature"

	// http header key
	ContentTypeHeader   = "Content-Type"
	ETagHeader          = "ETag"
	ContentLengthHeader = "Content-Length"

	// header value
	ContentTypeXMLHeaderValue = "application/xml"
)
