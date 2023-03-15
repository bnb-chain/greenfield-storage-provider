# Put Object

## Gateway
* Receives the PutObject request from the client.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account has permissions on resources.
* Dispatches the request to Uploader.

## Uploader
* Accepts object data in streaming and chops it into segments according to MaxSegmentSize. The MaxSegmentSize is the consensus result reached in the greenfield chain. And uploads the segments to PieceStore.
* Creates JobContext with the `INIT_UNSPECIFIED` initial state. Turns to `UPLOAD_OBJECT_DOING` state at the beginning of uploading segments. After uploading all segments, the JobContext's state enters `UPLOAD_OBJECT_DONE`. If any abnormal situation in the uploading, the JobContext's state will change to `UPLOAD_OBJECT_ERROR`.
* After uploading all the segments, insert all the segment data checksums and the root checksum into SP DB.
* Notifying the TaskNode, the Uploader will return to the client that the put object request is successful.

## TaskNode
* Asynchronously executes replicating object data to secondary SPs, and the uploader can always quickly receive the successful result from the TaskNode. The JobContext's state turn to `ALLOC_SECONDARY_DOING` from `UPLOAD_OBJECT_DONE`.
* Sends the GetSecondarySPApproval request to P2P node, it will broadcast to other SPs , and collect results back to TaskNode for selecting the secondary SPs. The JobContext's state enters `ALLOC_SECONDARY_DONE`, and turns into `REPLICATE_OBJECT_DOING` state immediately from `ALLOC_SECONDARY_DONE` state.
* Gets segments from PieceStore in parallel and computes a data redundancy solution for these segments based on Erasure Coding (EC), generating the EC pieces. Reorganize the EC pieces into six replicate data groups, each replicate data group contains several EC pieces according to the Redundancy policy.
* Then sends the replicate data groups in streaming to the selected secondary SPs in parallel.
* The secondary SP information of JobContext will be updated once if the replicating of a secondary SP is completed, until all secondary SPs are completed, the state of the JobContext will be updated to `REPLICATE_OBJECT_DONE` from `REPLICATE_OBJECT_DOING`.

## Receiver
* Checks the SecondarySP approval whether is self-signed and has timed out. If so, will return `SIGNATURE_ERROR` to the TaskNode.
* The Receiver works in the secondary SP, receives EC pieces that belong to the same replicate data group, and uploads the EC pieces to the secondary SP's PieceStore.
* Computes the EC pieces integrity checksum, sign the integrity checksum by SP's approval private key, then returns these to the TaskNode.

## TaskNode
* Receives the response from secondary SPs' Receiver, and un-sign the signature to compare with the secondary SP's approval public key.
* Sends the MsgSealObject to the Signer for signing the seal object transaction and broadcasting to the greenfield chain with the secondary SPs' integrity hash and signature. The state of the JobContext turns to `SIGN_OBJECT_DOING` from `REPLICATE_OBJECT_DONE`, if the Signer success to broadcast the SealObjectTX, then enters the `SIGN_OBJECT_DONE` state, and enters `SEAL_OBJECT_TX_DOING` state immediately from `SIGN_OBJECT_DONE` state.
* Monitor the execution results of seal object transaction on the greenfield chain to determine whether the seal is successful. If so, the JobContext state enters the `SEAL_OBJECT_DONE` state.


#### Background
* [PieceStore](../modules/01-piece_store.md)
* [Redundancy](../modules/02-redundancy.md)
* [JobContext](../modules/03-sp_db.md)
