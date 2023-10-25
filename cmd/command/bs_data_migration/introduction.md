# Introduction

The files in this directory are used to migrate block syncer data, usually required by a certain new feature or a bug fixing.

## Folder
The name of the folder represents the sp version of the repair scripts that the SP maintainer needs to run.

## migration command

All migration cmds under bs_data_migration folder should be idempotent.

### Example
The folder v1.0.1 contains a fix_payment job. This job was introduced in sp version v1.0.1. 
When SP maintainer upgrades the SP version to v1.0.1, he/she should execute the jobs defined in the folder v1.0.1.

The detailed job execution cmds could be found in migration_cmd_list.md.

## How to execute
1. Determine the target sp version and find the corresponding cmds in migration_cmd_list.md file
2. In SP running time (it could be a k8s container or a linux where the SP services are running on), you can run: `gndf-sp migration-config {{config.toml}} -j {{job_name}}`
   - config.toml is the configuration file path. 
   - job_name is the name of the script that needs to be run for this version.
