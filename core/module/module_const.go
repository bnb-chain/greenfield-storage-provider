package module

import (
	"strings"
)

var (
	ApprovalModularName              = strings.ToLower("Approver")
	ApprovalModularDescription       = "Handles the ask crate bucket/object and replicates piece approval request."
	AuthenticationModularName        = strings.ToLower("Authenticator")
	AuthenticationModularDescription = "Checks authentication."
	DownloadModularName              = strings.ToLower("Downloader")
	DownloadModularDescription       = "Downloads object and gets challenge info and statistical read traffic from the backend."
	ExecuteModularName               = strings.ToLower("TaskExecutor")
	ExecuteModularDescription        = "Executes background tasks."
	GateModularName                  = strings.ToLower("Gateway")
	GateModularDescription           = "Receives the user request and routes to the responding service."
	ManageModularName                = strings.ToLower("Manager")
	ManageModularDescription         = "Manages SPs and schedules tasks."
	P2PModularName                   = strings.ToLower("p2p")
	P2PModularDescription            = "Communicates between SPs on p2p protocol."
	ReceiveModularName               = strings.ToLower("Receiver")
	ReceiveModularDescription        = "Receives data pieces of an object from other storage provider and store."
	SignModularName                  = strings.ToLower("Signer")
	SignModularDescription           = "Signs the transaction and broadcasts to chain."
	UploadModularName                = strings.ToLower("Uploader")
	UploadModularDescription         = "Uploads object payload to primary SP."
)
