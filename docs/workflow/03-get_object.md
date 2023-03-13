# Get Object

## Gateway 
* Receives the GetObject request from the client.
* Verifies the signature of request to ensure that the request has not been tampered with.
* Checks the authorization to ensure the corresponding account has permissions on resources.
* Checks the object state and payment account state to ensure the object is sealed and payment account is active.
* Dispatches the request to Downloader.

## Downloader
* Check whether the read traffic exceeds the quota.
  * If exceeds the quota, the Downloader refuses to serve and returns a not-enough-quota response.
  * If smaller than quota, the Downloader read object data from the piece store, insert read record into the SP traffic-db and return the payload response.