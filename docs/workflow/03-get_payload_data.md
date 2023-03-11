# Get Object

## Gateway 
* Receives the GetObject request from the client.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account has permissions on resources.
* Checks the object state and payment account state to ensure the object is sealed and the payment account is active.
* Dispatches the request to Downloader.

## Downloader
* Receives the GetObject request from the Gateway service.
* Check whether the read traffic exceeds the quota.
  * If exceeds the quota, the Downloader refuses to serve and returns a not-enough-quota error to the Gateway.
  * If the quota is sufficient, the Downloader insert read record into the SP traffic-db.
* Splits the GetObject request info the GetPiece requests(support range read) to get piece payload data, and returns the object payload data streaming to the Gateway.