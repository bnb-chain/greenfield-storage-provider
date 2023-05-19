package module

import "strings"

var (
	ApprovalModularName                 = strings.ToLower("Approval")
	ApprovalModularDescription          = "Handles the ask crate bucket/object and replicate piece approval request."
	AuthorizationModularName            = strings.ToLower("Authorizer")
	AuthorizationModularDescription     = "Authorization authentication."
	DownloadModularName                 = strings.ToLower("Downloader")
	DownloadModularDescription          = "Downloads object and get challenge info from the backend and statistical read traffic."
	ExecuteModularName                  = strings.ToLower("TaskExecutor")
	ExecuteModularDescription           = "Executes background task."
	GateModularName                     = strings.ToLower("Gateway")
	GateModularDescription              = "Receives the user request and route to the responding service."
	ManageModularName                   = strings.ToLower("Manager")
	ManageModularDescription            = "SP management and task scheduling."
	P2PModularName                      = strings.ToLower("p2p")
	P2PModularDescription               = "Communicates between SPs on p2p protocol."
	ReceiveModularName                  = strings.ToLower("Receiver")
	ReceiveModularDescription           = "Receives data pieces of an object from other storage provider and store."
	SignerModularName                   = strings.ToLower("Signer")
	SignerModularDescription            = "Sign the transaction and broadcast to chain"
	UploadModularName                   = strings.ToLower("Uploader")
	UploadModularDescription            = "Uploads object payload to primary SP"
	BlockSyncerModularName              = strings.ToLower("BlockSyncer")
	BlockSyncerModularDescription       = "Synchronize data on the chain to SP"
	BlockSyncerModularBackupName        = strings.ToLower("BlockSyncerBackup")
	BlockSyncerModularBackupDescription = "Synchronize data on the chain to SP"
)
