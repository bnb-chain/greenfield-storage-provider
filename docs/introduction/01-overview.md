# Overview

## What is the Greenfield Storage Provider

Storage Providers (abbreviated SP) are storage service infrastructure providers.
They use Greenfield as the ledger and the single source of truth. Each SP can and
will respond to users' requests to write (upload) and read (download) data, and 
serve as the gatekeeper for user rights and authentications.

## Architecture

<div align=center><img src="../asset/01-sp_arch.jpg" alt="architecture.png" width="700"/></div>
<div align="center"><i>Storage Provider Architecture</i></div>

- **Gateway** is the entry point of each SP. It parses requests from the  client and dispatches them to special service.

- **Uploader** receives the object's payload data, splits it into segments, and stores them in piece store.

- **Downloader** handles the user's downloading request and gets object data from the piece store.

- **Receiver** receives data pieces from Primary SP and stores them in the piece store when SP works as a secondary SP.

- **Challenge** handles HA challenge requests and returns the challenged piece data and other pieces' hashes of the object.

- **TaskNode** works as the execute unit, it watches tasks(the smallest unit of a job) and executes them.

- **Manager** responsible for the service management of SP.

- **Signer** signs the transaction messages to the  Greenfield chain with the SP's private key.

- **P2P**  used to interact with the control flow of the payload data, eg: GetSecondaryApproval.

- **Metadata**  used to provide efficient query interface to achieve low latency and high-performance SP requirements.

- **PieceStore** interacts with underlying storage vendors, eg. AWS S3, MinIO.

- **SPDB** stores all the contexts of the background jobs and the metadata of the SP.

- **BSDB** stores all the events' data from the greenfield chain and provides them to the metadata service of SP.