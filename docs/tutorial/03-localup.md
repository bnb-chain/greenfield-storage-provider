## Setup Local StorageProviders

1. Generate localup env

Including build sp binary, generate directories/configs, create databases.
```bash
# The first time is equal to 1, and the other time is equal to 0.
# When equal to 1, the configuration template will be generated.
FIRST_TIME=1
bash ./deployment/localup/localup.sh --reset ${FIRST_TIME}
```

2. Overwrite db and sp info

Overwrite all sps' db.info and sp.info according to the real environment.

```
deployment/localup/local_env/
├── sp0
│   ├── config.toml
│   ├── db.info       # Overwrite db info
│   ├── gnfd-sp0
│   └── sp.info       # Overwrite sp info
├── sp1
├── ...
```

3. Start sp

Make config.toml real according to db.info and sp.info, and start sps.

```bash
# The first time is equal to 1, and the other time is equal to 0.
# When equal to 1, the configuration template will be generated.
FIRST_TIME=0
bash ./deployment/localup/localup.sh --reset ${FIRST_TIME}
```
The environment directory is as follows:
```
deployment/localup/local_env/
├── sp0
│   ├── config.toml
│   ├── data
│   ├── db.info
│   ├── gnfd-sp0
│   ├── gnfd-sp.log   # gnfd log file
│   ├── log.txt
│   └── sp.info
├── sp1
├── ...
```
4. Other supported commands

```bash
% bash ./deployment/localup/localup.sh --help
Usage: deployment/localup/localup.sh [option...] {help|reset|start|stop|print}

   --help                 display help info
   --reset $FIRST_TIME    reset env, $FIRST_TIME=0 or =1
   --start                start storage providers
   --stop                 stop storage providers
   --print                print sp local env work directory
```