package module

import (
	"strings"
)

var (
	ApprovalModularName             = strings.ToLower("Approval")
	ApprovalModularDescription      = "Handles the ask crate bucket/object and replicates piece approval request."
	AuthorizationModularName        = strings.ToLower("Authorizer")
	AuthorizationModularDescription = "Checks authorizations."
	DownloadModularName             = strings.ToLower("Downloader")
	DownloadModularDescription      = "Downloads object and gets challenge info and statistical read traffic from the backend."
	ExecuteModularName              = strings.ToLower("TaskExecutor")
	ExecuteModularDescription       = "Executes background tasks."
	GateModularName                 = strings.ToLower("Gateway")
	GateModularDescription          = "Receives the user request and routes to the responding service."
	ManageModularName               = strings.ToLower("Manager")
	ManageModularDescription        = "Manages SPs and schedules tasks."
	P2PModularName                  = strings.ToLower("p2p")
	P2PModularDescription           = "Communicates between SPs on p2p protocol."
	ReceiveModularName              = strings.ToLower("Receiver")
	ReceiveModularDescription       = "Receives data pieces of an object from other storage provider and store."
	SignerModularName               = strings.ToLower("Signer")
	SignerModularDescription        = "Signs the transaction and broadcasts to chain."
	UploadModularName               = strings.ToLower("Uploader")
	UploadModularDescription        = "Uploads object payload to primary SP."
)
