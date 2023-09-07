#!/usr/bin/env bash

export CGO_CFLAGS="-O -D__BLST_PORTABLE__"
export CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

MYSQL_USER="root"
MYSQL_PASSWORD="root"
MYSQL_ADDRESS="127.0.0.1:3306"
TESTCOVERAGE_THRESHOLD=60

workspace=${GITHUB_WORKSPACE}

function make_config() {
  cd ${workspace}
  make install-tools
  make build
  ./build/gnfd-sp config.dump
  cp config.toml ${workspace}/modular/blocksyncer/config.toml
  cd ${workspace}/modular/blocksyncer/ || exit 1

   # db
  sed -i -e "s/User = '.*'/User = '${MYSQL_USER}'/g" config.toml
  sed -i -e "s/Passwd = '.*'/Passwd = '${MYSQL_PASSWORD}'/g" config.toml
  sed -i -e "s/^Address = '.*'/Address = '${MYSQL_ADDRESS}'/g" config.toml
  sed -i -e "s/Database = '.*'/Database = 'block_syncer'/g" config.toml

  # chain
  sed -i -e "s/ChainID = '.*'/ChainID = 'greenfield_9000-1741'/g" config.toml
  sed -i -e "s/ChainAddress = \[.*\]/ChainAddress = \['http:\/\/localhost:8080'\]/g" config.toml

  # blocksyncer
  sed -i -e "s/Modules = \[\]/Modules = \[\'epoch\',\'bucket\',\'object\',\'payment\',\'group\',\'permission\',\'storage_provider\'\,\'prefix_tree\'\,\'virtual_group\'\,\'sp_exit_events\'\,\'object_id_map\'\]/g" config.toml
  WORKERS=10
  sed -i -e "s/Workers = 0/Workers = ${WORKERS}/g" config.toml

  echo "succeed to make config"
}

function reset_db() {
  hostname="localhost"
  port="3306"
  DATABASE="block_syncer"
  mysql -u ${MYSQL_USER} -h ${hostname} -P ${port} -p${MYSQL_PASSWORD} -e "drop database if exists ${DATABASE}"
  mysql -u ${MYSQL_USER} -h ${hostname} -P ${port} -p${MYSQL_PASSWORD} -e "create database ${DATABASE}"
}

function test_bs() {
  cd ${workspace}/modular/blocksyncer/ || exit 1
  go test -v -coverprofile=coverage.txt -covermode=atomic
  go tool cover -func coverage.txt

  echo "Quality Gate: checking test coverage is above threshold ..."
  echo "Threshold             : ${TESTCOVERAGE_THRESHOLD} %"
  totalCoverage=`go tool cover -func=coverage.txt | grep total | grep -Eo '[0-9]+\.[0-9]+'`
  echo "Current test coverage : $totalCoverage %"
  if (( $(echo "$totalCoverage ${TESTCOVERAGE_THRESHOLD}" | awk '{print ($1 > $2)}') )); then
      echo "OK"
  else
      echo "Current test coverage is below threshold. Please add more unit tests or adjust threshold to a lower value."
      echo "Failed"
      exit 1
  fi
}

function main() {
  CMD=$1
  case ${CMD} in
  --makecfg)
    make_config
    ;;
  --reset)
    reset_db
    ;;
  --start_test)
    test_bs
    ;;
  esac
}

main $@