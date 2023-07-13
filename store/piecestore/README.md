# PieceStore

## Requirement
- Basic API service, support P2P, RPC and HTTP
- Data cache, small pieces aggregation and GC
- Common interface for S3, Ceph, Minio, OSS, etc
- Sharding: multi buckets or multi regions
- Data profile
- Object storage benchmark

## CheckList

- [ ] s3: AWS S3
- [ ] file: local file, using disk persistence
- [ ] memory: memory storage, if server reboot, no data in disk
- [ ] minio: MinIO

## Usage

Amazon S3 supports two styles of [endPoint URI](https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/userguide/VirtualHosting.html): `path-style` and `virtual hosted-style`.
- Path-style: `https://s3.<region>.amazonaws.com/<bucket>`
- Virtual-hosted-style: `https://<bucket>.s3.<region>.amazonaws.com`

The <region> should be replaced with specific region code, e.g. the region code of US East (N. Virginia) is `us-east-1`. All the available region codes can be found [here](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions).

#### Note

For AWS users in China, you need add `.cn` to the host, i.e. `amazonaws.com.cn`, and check [this document](https://docs.amazonaws.cn/en_us/aws/latest/userguide/endpoints-arns.html) for region code.

### Permant credentials

Users can get `accessKey` and `secretKey` which used to verify users' identity from an object storage provider.

Public clouds typically allow users to create IAM (Identity and Access Management) roles, such as [AWS IAM role](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html), which can be assigned to VM instances. If the cloud server instance already has read and write access to the object storage, there is no need to specify `accessKey` and `secretKey`.

### Temporary access credentials

Other than permant credentials, users can also use temporary credentilas to access object storage through `accessKey`, `secretKey` and `sessionToken`. Temporary credentials have expired time, usually a few hours.

#### How to get temporary credentials?

Amazon S3 can refer this [link](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html) to get get temporary credentials.

### Sharding

The number of sharding in object storage that supports multi-bucket storage.

## Config Note

For safety, access key, secret key nad session token should be configured in environment:

```shell
export AWS_ACCESS_KEY="ACCESSKEY"
export AWS_SECRET_KEY="SECRETKEY"
export AWS_SESSION_TOKEN="SESSIONTOKEN"
```

BucketURL can be configured in either environment or config.toml.

```shell
export BUCKET_URL="BUCKETURL"
```

If BucketURL is configured in environment, all services will use the same bucket ro write and read data.

If `Shards` is not set in config.toml, the shard is 0, PieceStore won't shard.

> More storage providers will be supported
