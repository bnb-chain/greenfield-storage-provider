## Deployment

### Dependence

* SQL

    * no special requirements for the SQL DB version

      > MariaDB - 5.5.68 and Aurora(MySQL 5.7) 2.10.3 has been practiced.

* Payload Store

    * [AWS S3](https://aws.amazon.com/cn/s3/)
    * [MinIO](https://min.io/)(Beta)

* Golang

    * go1.19

### Compile

```shell
git clone https://github.com/bnb-chain/greenfield-storage-provider.git
cd greenfield-storage-provider && bash build.sh
cd build

# show the gnfd-sp version information
./gnfd-sp -v 

# show the gnfd-sp help
./gnfd-sp -h
```

### Setup

#### Join greenfield chain

// TODO:: waiting for the greenfield chain doc pr merged

#### Make configuration of greenfield storage provider

* Dump configuration template

  ```shell
  ./gnfd-sp config.dump
  ```

  A configuration template is being written to './config.toml'

* Edit configuration template

  // TODO:: the config file will change the format after changing will commit the template

### Start gnfd-sp with local model

```shell
./gnfd-sp -config ${config_file} -server ${service_name_list}
```

gnfd-sp supports any combination of services that are included in greenfield storage provider to run inside a process.

```shell
# show greenfield storeage field supports the services list 
./gnfd-sp list
```

Considering security factors, gnfd-sp supports obtaining sensitive information from environment variablesm includes:

```shell
AWS_ACCESS_KEY
AWS_SECRET_KEY
AWS_SESSION_TOKEN
BUCKET_URL

SP_DB_USER
SP_DB_PASSWORD

# signer service exclusive configuration
SIGNER_OPERATOR_PRIV_KEY
SIGNER_FUNDING_PRIV_KEY
SIGNER_APPROVAL_PRIV_KEY
SIGNER_SEAL_PRIV_KEY
```

The above information can be set to default, gnfd-sp will lookup from ENV if fields are the default value.

#### Start gnfd-sp with remote mode

The remote mode will upload the configuration to the SQL DB, avoid the inconsistency caused by configuration transfer.

TODO::support configuration dynamic delivery and hot loading.

* Upload configuration

  ```shell
  ./gnfd-sp config.upload -db.user ${db_user} -db.password ${db_password} -db.address ${db_address} -file ${config_file}
  
  # db,user and db.password flags support ENV Vars
  export SP_DB_USER=${SP_DB_USER}
  export SP_DB_PASSWORD=${SP_DB_PASSWORD}
  ./gnfd-sp config.upload -db.address ${db_address}
  
  ```

  The configuration will be uploaded to SQL

* Start service

  ```shell
  export SP_DB_USER=${SP_DB_USER}
  export SP_DB_PASSWORD=${SP_DB_PASSWORD}
  ./gnfd-sp config.remote -db.address ${db_address} -server ${service_name_list}
  ```

  

  

  