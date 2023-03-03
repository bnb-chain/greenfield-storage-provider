## Setup Local StorageProviders

1. Build Binary
```bash
make build
```

2. Dump Config Template
```bash
./build/gnfd-sp config.dump
```

3. Creates all the SP configuration files
```
.                                   # ~/.gnfd-sp
  |- localup_config/
      |- chain.info                 # chain-related configuration file.
      |- sp0.info                   # sp0-related configuration file, include priavate-keys and spdb-related configs.
      |- sp1.info                   
      |- sp2.info             
      |- sp3.info         
      |- sp4.info                   
      |- sp5.info             
      |- sp6.info         

```
All these configuration files are in ~/.gnfd-sp by default, but you need overwrite by the dependent chain and db.

4. Run local sps

Start all sps establish a storage network.
```bash
bash ./deployment/localup/localup.sh
```
The environment directory is as follows:
```
.                                   # ~/.gnfd-sp
  |- localup_env/
      |- sp0                        # ~/.sp0 work-directory
          |- gfsp                   # sp0 executable binary
          |- config.toml            # sp0 configuration file
          |- piece-store            # sp0 piece-store directory
          |- log.output             # sp0 log output
          
      |- sp1
      ...
```
Stop all sps.
```bash
bash ./deployment/localup/localup.sh stop
```
Clear all sps' spdb tables and local piece-store directories.
```bash
bash ./deployment/localup/localup.sh clear
```