package gateway

const (
	// query key
	TransactionQuery = "transaction"

	// bfs header key
	BFSRequestIDHeader       = "x-bfs-request-id"
	BFSContentLengthHeader   = "x-bfs-content-length"
	BFSContentTypeHeader     = "x-bfs-content-type"
	BFSCheckSumHeader        = "x-bfs-checksum"
	BFSIsPrivateHeader       = "x-bfs-is-private"
	BFSTransactionHashHeader = "x-bfs-transaction-hash"
	TransactionHeader        = "TransactionValue"

	// http header key
	ContentTypeHeader   = "Content-Type"
	ContentLengthHeader = "Content-Length"
	ETagHeader          = "ETag"

	// header value
	ContentTypeXMLHeaderValue = "application/xml"
)
