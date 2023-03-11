# Overview

## What is the Greenfield Storage Provider

Storage Providers (SP) are storage service infrastructures that organizations or individuals provide
and the corresponding roles they play. They use Greenfield as the ledger and the single source of truth. Each SP can and
will respond to users' requests to write (upload) and read (download) data, and serve as the gatekeeper for user rights and
authentications.

## Architecture
[architecture.png](!docs/asset/architecture.png)

### Gateway
The Gateway is the only service that interacts with the client, it receives the clients' 
requests and dispatches them to the Backend Service by RPC. Most client requests are 
HTTP requests following the S3 protocol.

### Uploader
The Uploader is responsible for uploading payload data, it's responsible for storing
data to Primary SP, and synchronizing to Secondary SP for Background Jobs completion.

### Downloader
The Downloader is responsible for reading payload and traffic billing, only Primary 
SP Data is responsible for responding to user GetObjet requests.

## Challenge
The Challenge is responsible for the challenger's(user or internal model) request, it
will return the payload integrity hash for the challenger to check whether the payload 
data is correctly stored.

### Receiver
The Receiver is a model unique to secondary SP, it's responsible for receiving a copy of
payload from the primary SP and will compute the payload integrity hash of the secondary 
SP and store it in the meta DB.

### Signer
The Signer is a singleton, holds and uses the SP's private key to sign the transaction 
then broadcasts to the Greenfield, Eg. SealObjectTX after uploading payload.

### Manager
The Manager is responsible for the service management of SP.

### TaskNode
The TaskNode is responsible for executing the background  job, such as syncing the primary SP's 
payload to the secondary SP.

### PieceStore
The module interacts with underlying storage vendors, with vendor-agnostic, production-
ready, high-performance, and other characteristics, Eg. AWS S3.

### Job DB
Store the context of background jobs, and SP meta data.