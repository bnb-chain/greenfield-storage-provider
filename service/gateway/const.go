package gateway

const (
	// path
	AdminPath          = "/greenfield/admin/v1"
	GetApprovalSubPath = "get-approval"

	// query key
	TransactionQuery = "transaction"
	PutObjectV2Query = "putobjectv2"
	ActionQuery      = "action"

	// bfs header key
	BFSRequestIDHeader       = "x-bfs-request-id"
	BFSContentLengthHeader   = "x-bfs-content-length"
	BFSContentTypeHeader     = "x-bfs-content-type"
	BFSChecksumHeader        = "x-bfs-checksum"
	BFSIsPrivateHeader       = "x-bfs-is-private"
	BFSTransactionHashHeader = "x-bfs-transaction-hash"
	BFSResourceHeader        = "x-bfs-resource"
	BFSPreSignatureHeader    = "x-bfs-pre-signature"

	// http header key
	ContentTypeHeader = "Content-Type"
	ETagHeader        = "ETag"
	// ContentLengthHeader = "Content-Length"

	// header value
	ContentTypeXMLHeaderValue = "application/xml"
)
