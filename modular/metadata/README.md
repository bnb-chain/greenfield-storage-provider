# Metadata
Metadata service is to supply better query service for the Inscription network. Users can interact with SP for some complex query services.
Some interfaces can be costly to implement on the chain or can cause significant latency.
metadata service is designed to implement the corresponding interface under the chain and provide it to the SP to achieve high performance and low latency.
The events' data are optimally stored by the block syncer and provided to the metadata.
Also, it provides additional extensions such as Pagination, Sort Key, and filtering. etc.

## Role
Sync all the Greenfield chain data to the distributed stores, and offers the read RPC
requests for chain data(in addition to payload). SP service will query the info, E.g.
permission, ListBucket, ListObject, etc. It will reduce the pressure on the Greenfield chain.

## Scalability
At present, the main role of metadata is to provide better scalability, and two main points are considered in the process of interface development:
1. the creation of interfaces that are not currently supported on the chain
2. metadata can provide better performance and low latency interfaces compared to those on the chain