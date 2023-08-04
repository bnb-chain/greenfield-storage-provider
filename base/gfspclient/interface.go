package gfspclient

// GfSpClientAPI for mock use
//
//go:generate mockgen -source=./interface.go -destination=./interface_mock.go -package=gfspclient
type GfSpClientAPI interface {
	ApproverAPI
	AuthenticatorAPI
	DownloaderAPI
	GaterAPI
	ManagerAPI
	MetadataAPI
	P2PAPI
	QueryAPI
	ReceiverAPI
	SignerAPI
	UploaderAPI
	Close() error
}
