# Deployment
## Dependence
* SQL: no special requirements for the SQL DB version
> MariaDB - 5.5.68 and Aurora(MySQL 5.7) 2.10.3 has been practiced.
* Payload Store: [AWS S3](https://aws.amazon.com/cn/s3/), [MinIO](https://min.io/)(Beta)

## Compile
```shell
# clone source
git clone https://github.com/bnb-chain/greenfield-storage-provider.git

# install complie tools
make install-tools 

# complie
cd greenfield-storage-provider && bash build.sh
cd build

# show the gnfd-sp version information
./gnfd-sp -v 

# show the gnfd-sp help
./gnfd-sp -h
```

## Join greenfield chain
> TODO:: waiting for the greenfield chain doc pr merged

## Make configuration
  ```shell
  # dump the configuration template to './config.toml'
  ./gnfd-sp config.dump
  
  # Edit configuration template
  ```

> TODO:: add the link of edit configuration example

## Start with local model
```shell
# show greenfield storeage field supports the services list 
./gnfd-sp list

# start 
./gnfd-sp -config ${config_file} -server ${service_name_list}
```

gnfd-sp supports obtaining sensitive information from environment variablesm includes:
```shell
# AWS
AWS_ACCESS_KEY
AWS_SECRET_KEY
AWS_SESSION_TOKEN
BUCKET_URL

# SQLDB
SP_DB_USER
SP_DB_PASSWORD

# signer service exclusive configuration
SIGNER_OPERATOR_PRIV_KEY
SIGNER_FUNDING_PRIV_KEY
SIGNER_APPROVAL_PRIV_KEY
SIGNER_SEAL_PRIV_KEY
```



## Start with remote mode
The remote mode will upload the configuration to the SQL DB, avoid the inconsistency caused by configuration transfer.
> TODO::support configuration dynamic delivery and hot loading.

  ```shell
  # Upload configuration
  ./gnfd-sp config.upload -db.user ${db_user} -db.password ${db_password} -db.address ${db_address} -file ${config_file}
  

  # start service
  export SP_DB_USER=${SP_DB_USER}
  export SP_DB_PASSWORD=${SP_DB_PASSWORD}
  ./gnfd-sp config.remote -db.address ${db_address} -server ${service_name_list}
  
  ```

  

  

  
