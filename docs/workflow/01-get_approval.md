# Get Approval

Before CreateBucket, PutObject, ReplicateData to SP, the request originator need send a GetApproval message to 
ask if it is willing to store the objects. The SP acknowledges the request by signing a message about the operation 
and returns it to the client; if the SP does not want to serve, it can refuse to sign.


## GetApproval to the primary SP
* The primary SP receive the CreateBucket/CreateObject GetApproval msg from the client.
  * If refuse to store, the primary SP send a refused response to the client.
  * If willing to store, the primary SP add the expired-height field to the msg and sign it, and response to the client.
* The client receive the GetApproval response msg.
  * If refuse to store, the client will retry other SPs.
  * If willing to store, the client forwards the approval to the greenfield chain, and the chain will check expired-height and signature.

## GetApproval to the secondary SPs
* The secondary SP receive the ReplicateData GetApproval msg from the primary SP.
  * If refuse to store, the secondary SP send a refused response to the primary SP.
  * If willing to store, the secondary SP add the expired-time field to the msg and sign it, and response to the primary SP.
* The primary SP receive the GetApproval response msg.
  * If refuse to store, the primary SP will retry other SPs.
  * If willing to store, the primary SP forwards the replicate data to the secondary SPs, and the secondary SPs will check expired-time and signature.
