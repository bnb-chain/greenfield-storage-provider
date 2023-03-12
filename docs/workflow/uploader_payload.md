# Upload Payload

## Concept

### Gateway
* permission check

### Uploader
* Accept payload in streaming and spilt it into segments according to MaxSegmentSize. The MaxSegmentSize is the consensus result reached in the Greenfield chain.
* Store the segments to [PieceStore](), the segment key format: ${object_id}+s${segment_idx}.
* Create JobContext and records in [SQL DB](), the initial state of the JobContext is INIT_UNSPECIFIED. Turn to UPLOAD_OBJECT_DOING at the beginning of storing segments. After storing all segments, the JobContext's state turn to UPLOAD_OBJECT_DONE. If any abnormal situation in the middle causes the uploading fail, the JobContext's status will change to UPLOAD_OBJECT_ERROR.
* Notify TaskNode to synchronize payload to other SPs.
* After uploading the segments and notifying the TaskNode, it will directly return to the client that the payload upload is successful.

### TaskNode(async)
* Asynchronously execute the synchronous payload to other SPs, and the uploader can always quickly receive the successful result of the TaskNode.
* Send [GetApproval]() to P2P node, it will broadcast to other SPs, and collect results back to TaskNode.
* Read segments from [PieceStore]() in parallel and perform EC encoding. Organize the data EC encoded into replication payload ready to be sent to other SPs, according to the [Redundancy]() policy.
* Streaming the replication payload to other SPs in parallel. Every time one SP is completed, the secondary information of JobContext will be updated once.

### Receiver
* The Receiver works in a secondary SP, receives the replication payload, and stores the replication payload into [PieceStore](), the piece key format: ${object_id}+s${segment_idx}+p${ec_idx}.
* Calculate the replication payload integrity hash, sign the integrity hash by SP's approval private key, and return to the TaskNode.

### TaskNode
* Receive the response from other SP's Receiver, and check the signature by unsign the signature to compare with the SP's approval public key.
* JobContext's state change from REPLICATE_OBJECT_DOING to REPLICATE_OBJECT_DONE.
* Send the SealObjextTX to the Signer for signing and broadcasting to the Greenfield chain.
* Monitor the execution results of SealObjextTX on the Greenfield chain to determine whether the seal is successful.