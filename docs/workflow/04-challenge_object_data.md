# Challenge Object Data
It is always the first priority of any decentralized storage network to guarantee data integrity and availability.
SP use the challenge workflow ensure it. Each SP splits the object payload to segments, and store segment data to piece store and store segment checksum to SP DB. 
And the greenfield chain store root checksum, which is calculated according to all segment checksum.
When users want to challenge object data, they need to specify segment index. The SP returns segment's payload data and all segments checksum, the user recalculates the root checksum and compare it with the root checksum on the greenfield chain

## Gateway
* Receives the Challenge request from the client.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account has permissions on resources.
* Dispatches the request to Challenge service.

## Challenge
* Receives the Challenge request from the Gateway.
* Returns all segment data checksums and challenge segment data payload to the Gateway service.
  * Retrieve all segment data checksums from the SP DB.
  * Get the challenge segment data from the piece store.
