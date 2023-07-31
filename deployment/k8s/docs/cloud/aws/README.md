Greenfield Storage Provider Deployment Guide - AWS
==================================================

## Pre-requisites (we assume you already have the following infrastructure):
1. AWS account
2. AWS EKS already set up


## High Level Architecture
![1](imgs/aws-infra-app-component.png "AWS Infra and SP Components")


### Resources
#### IAM role

* Create a new role which will be used by SP K8S application.
![1](imgs/iam-k8s-role.png "IAM Role")

* Add S3 permission policy - This is where SP stores its user uploaded content.
![2](imgs/iam-k8s-role-s3.png "IAM Role S3")

* Add Secret Manager permission policy - K8S will retrieve secret from here as app parameters
![3](imgs/iam-k8s-role-sm.png "IAM Role Secret Manager")

* Bind K8S service account to this IAM role
![4](imgs/iam-k8s-role-trust-relationship.png "IAM Role Trust Relationship")


#### Database (RDS)

* Create RDS database and jot down the connection string, username and password.
![5](imgs/rds.png)
after RDS created, need to init DB by creating databse:
1. db storage_provider_db
2. db block_syncer
3. db block_syncer_backup


#### S3 Bucket

* Create S3 bucket
![6](imgs/rds.png)


#### Secret Manager

* Create secret and update secret value (example provided below)
![7](imgs/secret-manager.png)

* For how the secret value should look like, please see examples/aws.

