# Challenge Object Data
It is always the first priority of any decentralized storage network to guarantee data integrity and availability.
We use data challenge instead of storage proof to get better HA. There will be some data challenges to random 
pieces on greenfield chain continuously. And the SP,  which stores the challenged piece, uses the challenge workflow 
to response. Each SP splits the object payload data to segments, and store segment data to piece store and store 
segment checksum to SP DB.

## Gateway
* Receives the Challenge request from the client.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account has permissions on resources.
* Dispatches the request to Challenge.

## Challenge
* Receives the Challenge request from the Gateway.
* Returns all segment data checksums and challenge segment data payload to the Gateway service.
  * Retrieve all segment data checksums from the SP DB.
  * Get the challenge segment data from the piece store.
