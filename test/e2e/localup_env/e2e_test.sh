#!/usr/bin/env bash

#basedir=$(cd `dirname $0` || return; pwd)
workspace=${GITHUB_WORKSPACE}

# some constants
GREENFIELD_TAG="v0.2.2-alpha.2"
MYSQL_USER="root"
MYSQL_PASSWORD="root"
MYSQL_IP="127.0.0.1"
MYSQL_PORT="3306"
TEST_ACCOUNT_ADDRESS="0x76263999b87D08228eFB098F36d17363Acf40c2c"
TEST_ACCOUNT_PRIVATE_KEY="da942d31bc4034577f581057e4a3644404ac12828a84052f87086d508fdcf095"

#########################################
# build and start Greenfield blockchain #
#########################################
function greenfield_chain() {
  # build Greenfield chain
  echo ${workspace}
  cd ${workspace} || return
  git clone https://github.com/bnb-chain/greenfield.git
  git checkout ${GREENFIELD_TAG}
  make proto-gen & make build

  # start Greenfield chain
  bash ./deployment/localup/localup.sh all 1 7
  bash ./deployment/localup/localup.sh export_sps 1 7 > sp.json
}

####################################
# build and start Greenfield SP    #
####################################
function greenfield_sp() {
  cd ${workspace} || return
  make install-tools & make build
  bash ./deployment/localup/localup.sh --generate ${workspace}/greenfield/sp.json ${MYSQL_USER} ${MYSQL_PASSWORD} ${MYSQL_IP}:${MYSQL_PORT}
  bash ./deployment/localup/localup.sh --reset
  bash ./deployment/localup/localup.sh --start
  sleep 10
  tail -n 1000 deployment/localup/local_env/sp0/gnfd-sp.log
  ps -ef | grep gnfd-sp | wc -l
}

#####################################
# build Greenfield cmd and test SP  #
#####################################
function test_sp() {
  cd ${workspace} || return
  # build sp
  git clone https://github.com/bnb-chain/greenfield-cmd.git
  make build
  cd build || return

  # generate a keystore file to manage private key information
  touch key.txt & echo ${TEST_ACCOUNT_PRIVATE_KEY} > key.txt
  touch password.txt & echo "test_sp_function" > password.txt
  ./gnfd-cmd --home ./ keystore generate --privKeyFile key.txt --passwordfile test.txt

  # construct config.toml
  touch config.toml
  {
    echo rpcAddr = "http://localhost:26750"
    echo chainId = "greenfield_9000-121"
  } > config.toml

  # test SP function: ls sp info; put object; download object
  transfer_account
  ./gnfd-cmd -c ./config.toml --home ./keystore sp ls
  sleep 5
  ./gnfd-cmd -c ./config.toml bucket create gnfd://sp_e2e_test_bucket
  sleep 10
  ./gnfd-cmd -c ./config.toml object put --contentType "application/json" ${workspace}/greenfield-storage-provider/test/e2e/localup_env/testdata/example.json gnfd://sp_e2e_test_bucket
  sleep 16
}

######################################
# transfer some BNB to test accounts #
######################################
function transfer_account() {
  cd cd ${workspace}/build/bin || return
  ./gnfd tx bank send validator0 ${TEST_ACCOUNT_ADDRESS} 500000000000000000000BNB --home ${workspace}/greenfield/deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -y
}

function main() {
  CMD=$1
  case ${CMD} in
  --startChain)
    greenfield_chain
    ;;
  --startSP)
    greenfield_sp
    ;;
  --runTest)
    test_sp
    ;;
  esac
}

main $@
