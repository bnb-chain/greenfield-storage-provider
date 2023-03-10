## Setup Local StorageProviders

## Dependence
* SQLDB: MariaDB - 5.5.68 and Aurora(MySQL 5.7) 2.10.3 has been practiced.

## Setup local greenfield chain
[setup private netWork](https://github.com/bnb-chain/greenfield/blob/master/docs/tutorial/03-local-network.md)

## Add SP to greenfield chain
[add sp to greenfield chain](https://github.com/bnb-chain/greenfield/blob/fynn/doc/docs/tutorial/07-storage-provider.md)

## Setup local sps
1. Generate localup env

Including build sp binary, generate directories/configs, create databases.
```bash
# The first time setup GEN_CONFIG_TEMPLATE=1, and the other time is 0.
# When equal to 1, the configuration template will be generated.
GEN_CONFIG_TEMPLATE=1
bash ./deployment/localup/localup.sh --reset ${GEN_CONFIG_TEMPLATE}
```

2. Overwrite db and sp info

Overwrite all sps' db.info and sp.info according to the real environment.

```
deployment/localup/local_env/
├── sp0
│   ├── config.toml   # templated config
│   ├── db.info       # to overwrite real db info
│   ├── gnfd-sp0      # sp binary
│   └── sp.info       # to overwrite real sp info
├── sp1
├── ...
```

3. Start sp

Make config.toml real according to db.info and sp.info, and start sps.

```bash
# In first time setup GEN_CONFIG_TEMPLATE=1, and the other time is 0.
# When equal to 1, the configuration template will be generated.
GEN_CONFIG_TEMPLATE=0
bash ./deployment/localup/localup.sh --reset ${GEN_CONFIG_TEMPLATE}
bash ./deployment/localup/localup.sh --start
```
The environment directory is as follows:
```
deployment/localup/local_env/
├── sp0
│   ├── config.toml    # real config
│   ├── data           # piecestore data directory
│   ├── db.info
│   ├── gnfd-sp0
│   ├── gnfd-sp.log    # gnfd log file
│   ├── log.txt
│   └── sp.info
├── sp1
├── ...
```
4. Other supported commands

```bash
% bash ./deployment/localup/localup.sh --help
Usage: deployment/localup/localup.sh [option...] {help|reset|start|stop|print}

   --help                           display help info
   --reset $GEN_CONFIG_TEMPLATE     reset env, $GEN_CONFIG_TEMPLATE=0 or =1
   --start                          start storage providers
   --stop                           stop storage providers
   --print                          print sp local env work directory
```
