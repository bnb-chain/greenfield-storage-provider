# Changelog

## v0.0.1

IMPROVEMENT
* [\#65](https://github:com/bnb-chain/greenfield-storage-provider/pull/65) feat: gateway add verify signature
* [\#43](https://github:com/bnb-chain/greenfield-storage-provider/pull/43) feat(uploader): add getAuth interface
* [\#68](https://github:com/bnb-chain/greenfield-storage-provider/pull/68) refactor: add jobdb v2 interface, objectID as primary key
* [\#70](https://github:com/bnb-chain/greenfield-storage-provider/pull/70) feat: change index from create object hash to object id
* [\#73](https://github:com/bnb-chain/greenfield-storage-provider/pull/73) feat(metadb): add sql metadb
* [\#82](https://github:com/bnb-chain/greenfield-storage-provider/pull/82) feat(stone_node): supports sending data to different storage provider
* [\#66](https://github:com/bnb-chain/greenfield-storage-provider/pull/66) fix: adjust the dispatching strategy of replica and inline data into storage provider
* [\#69](https://github.com/bnb-chain/greenfield-storage-provider/pull/69) fix: use multi-dimensional array to send piece data and piece hash
* [\#101](https://github.com/bnb-chain/greenfield-storage-provider/pull/101) fix: remove tokens from config and use env vars to load tokens
* [\#83](https://github.com/bnb-chain/greenfield-storage-provider/pull/83) chore(sql): polish sql workflow

Build
* [\#105](https://github.com/bnb-chain/greenfield-storage-provider/pull/105) fix: add release action
* [\#104](https://github.com/bnb-chain/greenfield-storage-provider/pull/104) fix: fix Dockerfile entrypoint instruction
* [\#67](https://github.com/bnb-chain/greenfield-storage-provider/pull/67) ci: add commit lint, code lint and unit test ci files
* [\#85](https://github.com/bnb-chain/greenfield-storage-provider/pull/85) chore: add pull request template
* [\#87](https://github.com/bnb-chain/greenfield-storage-provider/pull/87) chore: add setup-test-env tool

## v0.0.1-alpha
This is the first release of the gnfd-sp, mainly:
1. Implement alpha service skeleton
