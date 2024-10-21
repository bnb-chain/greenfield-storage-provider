#!/usr/bin/env bash

export CGO_CFLAGS="-O -D__BLST_PORTABLE__"
export CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

workspace=${GITHUB_WORKSPACE}

# some constants
MECHAIN_TAG="master"
# mechain cmd tag name: v0.1.0
MECHAIN_CMD_TAG="feat-adapt-tags"
# mechain go sdk tag name: v1.0.0
MECHAIN_GO_SDK_TAG="feat-adapt-tags"
MYSQL_USER="root"
MYSQL_PASSWORD="root"
MYSQL_ADDRESS="127.0.0.1:3306"
TEST_ACCOUNT_ADDRESS=${ACCOUNT_ADDR}
TEST_ACCOUNT_PRIVATE_KEY=${PRIVATE_KEY}
echo "TEST_ACCOUNT_ADDRESS is ""$TEST_ACCOUNT_ADDRESS"
echo "TEST_ACCOUNT_PRIVATE_KEY is ""$TEST_ACCOUNT_PRIVATE_KEY"

BUCKET_NAME="spbucket"

#########################################
# build and start Mechain blockchain #
#########################################
function mechain_chain() {
  set -e
  # build Mechain chain
  echo "${workspace}"
  cd "${workspace}"
  git clone https://github.com/zkMeLabs/mechain.git
  cd mechain/
  git checkout ${MECHAIN_TAG}
  make proto-gen &
  make build

  # start Mechain chain
  bash ./deployment/localup/localup.sh all 1 8
  bash ./deployment/localup/localup.sh export_sps 1 8 >sp.json

  # transfer some azkme tokens
  transfer_account
}

#############################################
# transfer some azkme tokens to test accounts #
#############################################
function transfer_account() {
  set -e
  cd "${workspace}"/mechain/
  ./build/bin/gnfd tx bank send validator0 "${TEST_ACCOUNT_ADDRESS}" 500000000000000000000azkme --home "${workspace}"/mechain/deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -y
  sleep 2
  ./build/bin/gnfd q bank balances "${TEST_ACCOUNT_ADDRESS}" --node http://localhost:26750
}

#################################
# build and start Mechain SP #
#################################
function mechain_sp() {
  set -e
  cd "${workspace}"
  make install-tools
  make build
  bash ./deployment/localup/localup.sh --generate "${workspace}"/mechain/sp.json ${MYSQL_USER} ${MYSQL_PASSWORD} ${MYSQL_ADDRESS}
  bash ./deployment/localup/localup.sh --reset
  bash ./deployment/localup/localup.sh --start
  sleep 60
  ./deployment/localup/local_env/sp0/gnfd-sp0 update.quota --quota 5000000000 -c deployment/localup/local_env/sp0/config.toml
  ./deployment/localup/local_env/sp1/gnfd-sp1 update.quota --quota 5000000000 -c deployment/localup/local_env/sp1/config.toml
  ./deployment/localup/local_env/sp2/gnfd-sp2 update.quota --quota 5000000000 -c deployment/localup/local_env/sp2/config.toml
  ./deployment/localup/local_env/sp3/gnfd-sp3 update.quota --quota 5000000000 -c deployment/localup/local_env/sp3/config.toml
  ./deployment/localup/local_env/sp4/gnfd-sp4 update.quota --quota 5000000000 -c deployment/localup/local_env/sp4/config.toml
  ./deployment/localup/local_env/sp5/gnfd-sp5 update.quota --quota 5000000000 -c deployment/localup/local_env/sp5/config.toml
  ./deployment/localup/local_env/sp6/gnfd-sp6 update.quota --quota 5000000000 -c deployment/localup/local_env/sp6/config.toml
  ./deployment/localup/local_env/sp7/gnfd-sp7 update.quota --quota 5000000000 -c deployment/localup/local_env/sp7/config.toml
  tail -n 1000 deployment/localup/local_env/sp0/mechain-sp.log
  ps -ef | grep mechain-sp | wc -l
}

############################################
# build Mechain cmd and set cmd config  #
############################################
function build_cmd() {
  set -e
  cd "${workspace}"
  # build sp
  git clone https://github.com/zkMeLabs/mechain-cmd.git
  cd mechain-cmd/
  git checkout ${MECHAIN_CMD_TAG}
  make build
  cd build/

  # generate a keystore file to manage private key information
  touch key.txt &
  echo "${TEST_ACCOUNT_PRIVATE_KEY}" >key.txt
  touch password.txt &
  echo "test_sp_function" >password.txt
  ./gnfd-cmd --home ./ --passwordfile password.txt account import key.txt

  # construct config.toml
  touch config.toml
  {
    echo rpcAddr = \"http://localhost:26750\"
    echo chainId = \"mechain_5151-1\"
  } >config.toml
}

############################################
# build Mechain go-sdk                  #
############################################
function build_mechain-go-sdk() {
  set -e
  cd "${workspace}"
  # build mechain-go-sdk
  git clone https://github.com/zkMeLabs/mechain-go-sdk.git
  cd mechain-go-sdk/
  git checkout ${MECHAIN_GO_SDK_TAG}
}

######################
# test create bucket #
######################
function test_create_bucket() {
  set -e
  cd "${workspace}"/mechain-cmd/build/
  ./gnfd-cmd -c ./config.toml --home ./ sp ls
  sleep 5
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt bucket create gnfd://${BUCKET_NAME}
  ./gnfd-cmd -c ./config.toml --home ./ bucket head gnfd://${BUCKET_NAME}
  sleep 10
}

###########################################################
# test upload and download file which size less than 16MB #
###########################################################
function test_file_size_less_than_16_mb() {
  set -e
  cd "${workspace}"/mechain-cmd/build/
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object put --contentType "application/json" "${workspace}"/test/e2e/spworkflow/testdata/example.json gnfd://${BUCKET_NAME}
  sleep 32
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object get gnfd://${BUCKET_NAME}/example.json ./test_data.json
  check_md5 "${workspace}"/test/e2e/spworkflow/testdata/example.json ./test_data.json
  cat test_data.json
}

##############################################################
# test upload and download file which size greater than 16MB #
##############################################################
function test_file_size_greater_than_16_mb() {
  set -e
  cd "${workspace}"/mechain-cmd/build/
  dd if=/dev/urandom of=./random_file bs=17M count=1
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object put --contentType "application/octet-stream" ./random_file gnfd://${BUCKET_NAME}/random_file
  sleep 32
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object get gnfd://${BUCKET_NAME}/random_file ./new_random_file
  sleep 10
  check_md5 ./random_file ./new_random_file
}

################
# test sp exit #
################
function test_sp_exit() {
  set -xe
  # choose sp5
  cd "${workspace}"/deployment/localup/local_env/sp5
  operator_address=$(echo "$(grep "SpOperatorAddress" ./config.toml)" | grep -o "0x[0-9a-zA-Z]*")
  echo "${operator_address}"
  cd "${workspace}"/mechain-cmd/build/
  ls
  dd if=/dev/urandom of=./random_file bs=17M count=1
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt bucket create --primarySP "${operator_address}" gnfd://spexit
  ./gnfd-cmd -c ./config.toml --home ./ bucket head gnfd://spexit
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object put --contentType "application/octet-stream" ./random_file gnfd://spexit/random_file
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object put --contentType "application/json" "${workspace}"/test/e2e/spworkflow/testdata/example.json gnfd://spexit/example.json
  sleep 16
  ./gnfd-cmd -c ./config.toml --home ./ object head gnfd://spexit/random_file
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object get gnfd://spexit/random_file ./new_random_file
  ./gnfd-cmd -c ./config.toml --home ./ object head gnfd://spexit/example.json
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object get gnfd://spexit/example.json ./new.json

  sleep 10
  check_md5 "${workspace}"/test/e2e/spworkflow/testdata/example.json ./new.json
  check_md5 ./random_file ./new_random_file

  # start exiting sp5
  cd "${workspace}"/deployment/localup/local_env/sp5
  ./gnfd-sp5 -c ./config.toml sp.exit -operatorAddress "${operator_address}"
  cd "${workspace}"/mechain-cmd/build/
  ./gnfd-cmd -c ./config.toml --home ./ sp ls
  sleep 180
  ./gnfd-cmd -c ./config.toml --home ./ sp ls
  ./gnfd-cmd -c ./config.toml --home ./ bucket head gnfd://spexit
  ./gnfd-cmd -c ./config.toml --home ./ object head gnfd://spexit/example.json
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object get gnfd://spexit/example.json ./new1.json
  ./gnfd-cmd -c ./config.toml --home ./ --passwordfile password.txt object get gnfd://spexit/random_file ./new_random_file1
  sleep 10
  check_md5 "${workspace}"/test/e2e/spworkflow/testdata/example.json ./new1.json
  check_md5 ./random_file ./new_random_file1
}

##################################
# check two md5 whether is equal #
##################################
function check_md5() {
  set -e
  if [ $# != 2 ]; then
    echo "failed to check md5 value; this function needs two args"
    exit 1
  fi
  file1=$1
  file2=$2
  md5_1=$(md5sum "${file1}" | cut -d ' ' -f 1)
  md5_2=$(md5sum "${file2}" | cut -d ' ' -f 1)
  echo "${md5_1}"
  echo "${md5_2}"

  if [ "$md5_1" = "$md5_2" ]; then
    echo "The md5 values are the same."
  else
    echo "The md5 values are different."
    exit 1
  fi
}

#######################
# run sp workflow e2e #
#######################
function run_e2e() {
  set -e
  echo 'run test_create_bucket'
  test_create_bucket
  echo 'run put object case less than 16 MB'
  test_file_size_less_than_16_mb
  echo 'run put object case greater than 16 MB'
  test_file_size_greater_than_16_mb
}

###################
# run sp exit e2e #
###################
# TODO: use this function in sp exit e2e for speeding all e2e process which will be overwritten in the future
function run_sp_exit_e2e() {
  set -e
  echo 'run sp exit e2e test'
  test_sp_exit
}

###################
# run go-sdk e2e #
###################
function run_go_sdk_e2e() {
  set +e
  cd "${workspace}"/mechain-go-sdk/
  echo 'run mechain go sdk e2e test'
  go test -v e2e/e2e_migrate_bucket_test.go
  exit_status_command=$?
  if [ $exit_status_command -eq 0 ]; then
    echo "make e2e_test successful."
  else
    cat "${workspace}"/deployment/localup/local_env/sp0/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp1/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp2/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp3/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp4/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp5/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp6/log.txt
    cat "${workspace}"/deployment/localup/local_env/sp7/log.txt
    exit $exit_status_command
  fi
}

function main() {
  CMD=$1
  case ${CMD} in
  --startChain)
    mechain_chain
    ;;
  --startSP)
    mechain_sp
    ;;
  --buildCmd)
    build_cmd
    ;;
  --runTest)
    run_e2e
    ;;
  --runSPExit)
    #    run_sp_exit_e2e
    ;;
  --runSDKE2E)
    build_mechain-go-sdk
    run_go_sdk_e2e
    ;;
  esac
}

main $@
