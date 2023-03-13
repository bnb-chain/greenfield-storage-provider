# Challenge Object Data
It is always the first priority of any decentralized storage network to guarantee data integrity and availability.
Greenfield use the challenge workflow ensure it.

## Gateway
* Receives the Challenge request from the client.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account has permissions on resources.
* Dispatches the request to Challenge service.

## Challenge
* Receives the Challenge request from the Gateway.
* Returns all segment data checksums and challenge segment data payload to the client.
  * Retrieve all segment data checksums from the SP DB.
  * Get the challenge segment data from the piece store.
