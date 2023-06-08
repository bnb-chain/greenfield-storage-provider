Greenfield Storage Provider Deployment Guide
============================================

Supported Cloud Providers
-------------------------
We currently support only deployment to AWS (EKS). Other cloud providers (e.g. Alicloud, GCP)
are in our pipeline and will be supported in the future.


Quick Start
-----------

For detail about what "storage provider" is on application level, please see
https://greenfield.bnbchain.org/docs/guide/storage-provider/. This document focuses on AWS infra
and K8S deployment level.


## Pre-requisites (we assume you already have the following infrastructure):
1. AWS account
2. AWS EKS available
3. K8S kustomize client
4. For monitoring (optional):
     1. Victoria Metrics
     2. Grafana dashboard
     3. Alert channels

## High Level Architecture
![1](../../asset/04-aws-infra-app-component.png "AWS Infra and SP Components")


## Steps:
1. [Create AWS resources](aws/)
2. [Create K8S resources](k8s/)
4. [Set up monitoring dashboard](grafana/)


