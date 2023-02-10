#!/bin/bash
set -e

exec "/app/gnfd-sp -config /app/config.toml > gnfd-sp.log 2>&1 &" "$@"
