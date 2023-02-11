# Changelog


## v0.0.1

IMPROVEMENT
* [\#65](https://github.com/bnb-chain/greenfield-storage-provider/pull/65) feat: gateway add verify signature
* [\#43](https://github.com/bnb-chain/greenfield-storage-provider/pull/43) feat(uploader): add getAuth interface
* [\#68](https://github.com/bnb-chain/greenfield-storage-provider/pull/68) refactor: add jobdb v2 interface, objectID as primary key
* [\#70](https://github.com/bnb-chain/greenfield-storage-provider/pull/70) feat: change index from create object hash to object id
* [\#73](https://github.com/bnb-chain/greenfield-storage-provider/pull/73) feat(metadb): add sql metadb
* [\#82](https://github.com/bnb-chain/greenfield-storage-provider/pull/82) feat(stone_node): supports sending data to different storage provider
* [\#66](https://github.com/bnb-chain/greenfield-storage-provider/pull/66) fix: adjust the dispatching strategy of replica and inline data into storage provider
* [\#69](https://github.com/bnb-chain/greenfield-storage-provider/pull/69) fix: use multi-dimensional array to send piece data and piece hash
* [\#101](https://github.com/bnb-chain/greenfield-storage-provider/pull/101) fix: remove tokens from config and use env vars to load tokens
* [\#83](https://github.com/bnb-chain/greenfield-storage-provider/pull/83) chore(sql): polish sql workflow
* [\#87](https://github.com/bnb-chain/greenfield-storage-provider/pull/87) chore: add setup-test-env tool

Build
* [\#74](https://github.com/bnb-chain/greenfield-storage-provider/pull/74) ci: add docker release pipe
* [\#67](https://github.com/bnb-chain/greenfield-storage-provider/pull/67) ci: add commit lint, code lint and unit test ci files
* [\#85](https://github.com/bnb-chain/greenfield-storage-provider/pull/85) chore: add pull request template
* [\#105](https://github.com/bnb-chain/greenfield-storage-provider/pull/105) fix: add release action


## v0.0.1-alpha
This release includes features, mainly:
1. Implement the upload and download of payload data and the challenge handler api of piece data;
2. Implement the main architecture of greenfield storage provider:  
   2.1 gateway: the entry point of each sp, parses requests from the client and dispatches them to special service;  
   2.2 uploader: receives the object's payload data, splits it into segments, and stores them in piece store;   
   2.3 downloader: handles the user's downloading request and gets object data from the piece store;    
   2.4 stonehub: works as state machine to handle all background jobs, each job includes several tasks;   
   2.5 stonenode: works as the execute unit, it watches the stonehub tasks(the smallest unit of a job) and executes them;   
   2.6 syncer: receives data pieces from primary sp and stores them in the piece store when sp works as a secondary sp;
3. Implement one-click deployment and one-click running test, which is convenient for developers and testers to experience the gnfd-sp.

* [\#7](https://github.com/bnb-chain/greenfield-storage-provider/pull/7) feat(gateway/uploader): add gateway and uploader skeleton
* [\#16](https://github.com/bnb-chain/greenfield-storage-provider/pull/16) Add secondary syncer service
* [\#17](https://github.com/bnb-chain/greenfield-storage-provider/pull/17) feat: implement of upload payload in stone hub side
* [\#29](https://github.com/bnb-chain/greenfield-storage-provider/pull/28) fix: ston node goroutine model
* [\#38](https://github.com/bnb-chain/greenfield-storage-provider/pull/38) feat: implement the challenge service
* [\#9](https://github.com/bnb-chain/greenfield-storage-provider/pull/9) add service lifecycle module
* [\#2](https://github.com/bnb-chain/greenfield-storage-provider/pull/2) add piecestore module
* [\#18](https://github.com/bnb-chain/greenfield-storage-provider/pull/18) feat: add job meta orm
* [\#60](https://github.com/bnb-chain/greenfield-storage-provider/pull/60) test: add run cases
