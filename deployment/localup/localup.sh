#!/usr/bin/env bash

basedir=$(cd `dirname $0`; pwd)
workspace=${basedir}
source ${workspace}/env.info
sp_bin_name=gnfd-sp
sp_bin=${workspace}/../../build/${sp_bin_name}

#########################
# the command line help #
#########################
display_help() {
    echo "Usage: $0 [option...] {help|reset|start|stop|print}" >&2
    echo
    echo "   --help                           display help info"
    echo "   --reset \$GEN_CONFIG_TEMPLATE     reset env, \$GEN_CONFIG_TEMPLATE=0 or =1"
    echo "   --start                          start storage providers"
    echo "   --stop                           stop storage providers"
    echo "   --print                          print sp local env work directory"
    echo
    exit 0
}

################################
# generate sp config templates #
################################
generate_env() {
  gen_config_template=$1
  bash ${workspace}/../../build.sh
  mkdir -p ${workspace}/${SP_DEPLOY_DIR}

  for ((i=0;i<${SP_NUM};i++));do
    mkdir -p ${workspace}/${SP_DEPLOY_DIR}/sp${i}
    cp -rf ${sp_bin} ${workspace}/${SP_DEPLOY_DIR}/sp${i}/${sp_bin_name}${i}
    if [ ${gen_config_template} -eq 1 ]
    then
      cd ${workspace}/${SP_DEPLOY_DIR}/sp${i}/
      ./${sp_bin_name}${i}  config.dump
      {
        echo '#!/usr/bin/env bash'
        echo 'USER=""'
        echo 'PWD=""'
        echo 'ADDRESS=""'
        echo 'DATABASE=""'
      } > db.info
      {
        echo '#!/usr/bin/env bash'
        echo 'SP_ENDPOINT=""'
        echo 'OPERATOR_ADDRESS=""'
        echo 'OPERATOR_PRIVATE_KEY=""'
        echo 'FUNDING_PRIVATE_KEY=""'
        echo 'SEAL_PRIVATE_KEY=""'
        echo 'APPROVAL_PRIVATE_KEY=""'
      } > sp.info
      cd - >/dev/null
    fi
  done
}

###############################################################
# make sp config.toml real according env.info/db.info/sp.info #
###############################################################
make_config() {
  index=0
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
      cur_port=$((SP_START_PORT+1000*$index))
      cd ${sp_dir}
        source db.info
        source sp.info
        # db
        sed -i -e "s/root/${USER}/g" config.toml
        sed -i -e "s/test_pwd/${PWD}/g" config.toml
        sed -i -e "s/localhost\:3306/${ADDRESS}/g" config.toml
        sed -i -e "s/storage_provider_db/${DATABASE}/g" config.toml
        # sp
        sed -i -e "s/localhost\:9033/${SP_ENDPOINT}/g" config.toml
        sed -i -e "s/9133/$(($cur_port+133))/g" config.toml
        sed -i -e "s/9233/$(($cur_port+233))/g" config.toml
        sed -i -e "s/9333/$(($cur_port+333))/g" config.toml
        sed -i -e "s/9433/$(($cur_port+433))/g" config.toml
        sed -i -e "s/9533/$(($cur_port+533))/g" config.toml
        sed -i -e "s/9633/$(($cur_port+633))/g" config.toml
        sed -i -e "s/9733/$(($cur_port+733))/g" config.toml
        sed -i -e "s/9833/$(($cur_port+833))/g" config.toml
        sed -i -e "s/9933/$(($cur_port+933))/g" config.toml
        sed -i -e "s/24036/$(($cur_port+4036))/g" config.toml
        sed -i -e "s/SpOperatorAddress = \".*\"/SpOperatorAddress = \"${OPERATOR_ADDRESS}\"/g" config.toml
        sed -i -e "s/OperatorPrivateKey = \".*\"/OperatorPrivateKey = \"${OPERATOR_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/FundingPrivateKey = \".*\"/FundingPrivateKey = \"${FUNDING_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/SealPrivateKey = \".*\"/SealPrivateKey = \"${SEAL_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/ApprovalPrivateKey = \".*\"/ApprovalPrivateKey = \"${APPROVAL_PRIVATE_KEY}\"/g" config.toml
        # chain
        sed -i -e "s/greenfield_9000-1741/${CHAIN_ID}/g" config.toml
        sed -i -e "s/localhost\:9090/${CHAIN_GRPC_ENDPOINT}/g" config.toml
        sed -i -e "s/localhost\:26750/${CHAIN_HTTP_ENDPOINT}/g" config.toml
        echo "succeed to generate config.toml in "${sp_dir}
        # p2p
        node=NODE${index}
        boot=BOOT${index}
        sed -i -e "s/P2PPrivateKey = \".*\"/P2PPrivateKey = \"${!node}\"/g" config.toml
        sed -i -e "s/Bootstrap = \[\]/Bootstrap = \[\"${!boot}\"\]/g" config.toml
      cd - >/dev/null
      index=$(($index+1))
  done
}

#############
# start sps #
#############
start_sp() {
  index=0
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
    cd ${sp_dir}
    nohup ./${sp_bin_name}${index} --config config.toml </dev/null >log.txt 2>&1&
    echo "succeed to start sp in "${sp_dir}
    cd - >/dev/null
    index=$(($index+1))
  done
  echo "succeed to start storage providers"
}

############
# stop sps #
############
stop_sp() {
  kill -9 $(pgrep -f ${sp_bin_name}) >/dev/null 2>&1
  echo "succeed to finish storage providers"
}

#############################################
# drop databases and recreate new databases #
#############################################
reset_db() {
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
    cd ${sp_dir}
    source db.info
    hostname=$(echo ${ADDRESS} | cut -d : -f 1)
    port=$(echo ${ADDRESS} | cut -d : -f 2)
    mysql -u ${USER} -h ${hostname} -P ${port} -p${PWD} -e "drop database if exists ${DATABASE}"
    mysql -u ${USER} -h ${hostname} -P ${port} -p${PWD} -e "create database ${DATABASE}"
    echo "succeed to reset db in "${sp_dir}
    cd - >/dev/null
  done
}

##########################
# clean piece-store data #
##########################
reset_store() {
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
      cd ${sp_dir}
      rm -rf ./data
      echo "succeed to reset store in "${sp_dir}
      cd - >/dev/null
    done
}

#############
# reset sps #
#############
reset_sp() {
  if [ $# != 1 ] ; then
    echo "failed to reset sp, please check args by help info"
    exit 1
  fi
  stop_sp
  gen_config_template=$1
  if [ ${gen_config_template} -eq 1 ]
  then
    generate_env ${gen_config_template}
    echo
    echo "succeed to generate templates, and need to overwrite db.info and sp.info in the following working directory:"
    print_work_dir
    echo "after overwrite, reset again with FIRST_TIME=0"
    echo
  else
    reset_db
    reset_store
    generate_env 0
    make_config
  fi
}

##################
# print work dir #
##################
print_work_dir() {
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
    echo "  "${sp_dir}
  done
}

main() {
  CMD=$1
  case ${CMD} in
  --reset)
    reset_sp $2
    ;;
  --start)
    stop_sp
    start_sp
    ;;
  --stop)
    stop_sp
    ;;
  --print)
    print_work_dir
    ;;
  --help|*)
    display_help
    ;;
  esac
}

main $@