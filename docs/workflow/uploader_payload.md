# Upload Payload

### Gateway
* permission check

### Uploader
* data stream 
* spilt segments
* piece store
* notify to task node

### TaskNode(async)
* parallel get segments from piece store
* get approval by p2p
* parallel EC encode segments to piece
* parallel sync pieces to secondary SPs

### Receiver
* check the approval
* receive the piece for net stream
* compute the integrity hash and return primary SP's TaskNode

### TaskNode
* receive the response from secondary SP's Receiver
* check the signature
* send seal object tx to singer

### Signer
* sign the seal object tx and broadcast to greenfield chain

### TaskNode
* check the object status by object id from the greenfield chain 