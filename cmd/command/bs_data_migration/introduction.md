# Introduction

The files in this directory are used to fix data problems caused by blocksyncer bugs

## Folder
The name of the folder represents the sp version of the repair script you need to run

### Explame
Current sp version: v1.0.0

You need to run the job under the folder fix-v1.0.0 to fix the data and subsequently update to the next version (if the next version exists).

Before upgrading the version, you must run the repair script first, otherwise the block processed in the current version will appear wrong data


## How to execute
1. Determine the current sp version and find the corresponding script
2. On pod, run: gndf-sp migration-config {{config.toml}} -j {{job_name}}

config.toml is the configuration file path

job_name is the name of the script that needs to be run for this version, and there may be several.

In the job_name file in the fix folder, we give the name of the script that needs to be run for each version
