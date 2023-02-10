#!/bin/bash
set -e

exec "/app/gnfd-sp -config config.toml > gnfd-sp.log 2>&1 &" "$@"
