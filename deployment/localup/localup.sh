#!/usr/bin/env bash

basedir=$(cd `dirname $0`; pwd)
workspace=${basedir}
source ${workspace}/env.info
sp_bin_name=gnfd-sp
sp_bin=${workspace}/../../build/${sp_bin_name}
gnfd_bin=${workspace}/../../../greenfield/build/bin/gnfd
gnfd_workspace=${workspace}/../../../greenfield/deployment/localup/

#########################
# the command line help #
#########################
display_help() {
    echo "Usage: $0 [option...] {help|generate|reset|start|stop|print}" >&2
    echo
    echo "   --help           display help info"
    echo "   --generate       generate sp.info and db.info that accepts four args: the first arg is json file path, the second arg is db username, the third arg is db password and the fourth arg is db address"
    echo "   --reset          reset env"
    echo "   --start          start storage providers"
    echo "   --stop           stop storage providers"
    echo "   --print          print sp local env work directory"
    echo
    exit 0
}

################################
# generate sp.info and db.info #
################################
function generate_sp_db_info() {
  if [ $# != 4 ] ; then
    echo "failed to generate sp.info and db.info, please check args by help info"
    exit 1
  fi
  bash ${workspace}/../../build.sh
  mkdir -p ${workspace}/${SP_DEPLOY_DIR}

  sp_json_file=$1
  db_user=$2
  db_password=$3
  db_address=$4
  for ((i=0;i<${SP_NUM};i++));do
    mkdir -p ${workspace}/${SP_DEPLOY_DIR}/sp${i}
    cp -rf ${sp_bin} ${workspace}/${SP_DEPLOY_DIR}/sp${i}/${sp_bin_name}${i}
    cd ${workspace}/${SP_DEPLOY_DIR}/sp${i}/
    ./${sp_bin_name}${i}  config.dump

    # generate sp info
    i_port=`expr ${SP_START_ENDPOINT_PORT} + $i`
    endpoint="127.0.0.1:${i_port}"
    {
      echo "#!/usr/bin/env bash"
      echo "SP_ENDPOINT=\"${endpoint}\""
    } > sp.info
    op_address=$(jq -r ".sp${i}.OperatorAddress" ${sp_json_file})
    echo "OPERATOR_ADDRESS=\"${op_address}\"" >> sp.info
    opk=$(jq -r ".sp${i}.OperatorPrivateKey" ${sp_json_file})
    echo "OPERATOR_PRIVATE_KEY=\"${opk}\"" >> sp.info
    fpk=$(jq -r ".sp${i}.FundingPrivateKey" ${sp_json_file})
    echo "FUNDING_PRIVATE_KEY=\"${fpk}\"" >> sp.info
    spk=$(jq -r ".sp${i}.SealPrivateKey" ${sp_json_file})
    echo "SEAL_PRIVATE_KEY=\"${spk}\"" >> sp.info
    apk=$(jq -r ".sp${i}.ApprovalPrivateKey" ${sp_json_file})
    echo "APPROVAL_PRIVATE_KEY=\"${apk}\"" >> sp.info
    gpk=$(jq -r ".sp${i}.GcPrivateKey" ${sp_json_file})
    echo "GC_PRIVATE_KEY=\"${gpk}\"" >> sp.info

    # generate db info
    {
      echo '#!/usr/bin/env bash'
      echo "USER=\"${db_user}\""
      echo "PWD=\"${db_password}\""
      echo "ADDRESS=\"${db_address}\""
      echo "DATABASE=sp_${i}"
    } > db.info
    cd - >/dev/null
  done
  print_work_dir
  echo "succeed to generate sp.info and db.info"
  echo
}

#############################################################
# make sp config.toml according to env.info/db.info/sp.info #
#############################################################
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
        sed -i -e "s/8933/$(($cur_port+33))/g" config.toml
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
        sed -i -e "s/25341/$(($cur_port+5341))/g" config.toml
        sed -i -e "s/SpOperatorAddress = \".*\"/SpOperatorAddress = \"${OPERATOR_ADDRESS}\"/g" config.toml
        sed -i -e "s/OperatorPrivateKey = \".*\"/OperatorPrivateKey = \"${OPERATOR_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/FundingPrivateKey = \".*\"/FundingPrivateKey = \"${FUNDING_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/SealPrivateKey = \".*\"/SealPrivateKey = \"${SEAL_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/ApprovalPrivateKey = \".*\"/ApprovalPrivateKey = \"${APPROVAL_PRIVATE_KEY}\"/g" config.toml
        sed -i -e "s/GcPrivateKey = \".*\"/GcPrivateKey = \"${GC_PRIVATE_KEY}\"/g" config.toml
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

#############################################################
# make intergation test config.toml according sp.json       #
#############################################################
make_intergation_test_config() {
  index=0
  sp_json_file=$1
  file='test/e2e/localup_env/integration_config/config.yml'

  validator_priv_key=("$(echo "y" | $gnfd_bin keys export validator0 --unarmored-hex --unsafe --keyring-backend test --home ${gnfd_workspace}/.local/validator0)")

  sed -i -e "s/20f92afe113b90e1faa241969e957ac091d80b920f84ffda80fc9d0588f62906/${validator_priv_key}/g" $file

  echo "SPs:" >> $file
  sp0_opk=$(jq -r ".sp0.OperatorPrivateKey" ${sp_json_file})
  sp0_fpk=$(jq -r ".sp0.FundingPrivateKey" ${sp_json_file})
  sp0_spk=$(jq -r ".sp0.SealPrivateKey" ${sp_json_file})
  sp0_apk=$(jq -r ".sp0.ApprovalPrivateKey" ${sp_json_file})

  sp0_opaddr=$(jq -r ".sp0.OperatorAddress" ${sp_json_file})
  sp1_opaddr=$(jq -r ".sp1.OperatorAddress" ${sp_json_file})
  sp2_opaddr=$(jq -r ".sp2.OperatorAddress" ${sp_json_file})
  sp3_opaddr=$(jq -r ".sp3.OperatorAddress" ${sp_json_file})
  sp4_opaddr=$(jq -r ".sp4.OperatorAddress" ${sp_json_file})
  sp5_opaddr=$(jq -r ".sp5.OperatorAddress" ${sp_json_file})
  sp6_opaddr=$(jq -r ".sp6.OperatorAddress" ${sp_json_file})


  echo "  - OperatorSecret: "${sp0_opk}"" >> $file
  echo "    FundingSecret: "${sp0_fpk}"" >> $file
  echo "    ApprovalSecret: "${sp0_spk}"" >> $file
  echo "    SealSecret: "${sp0_apk}"" >> $file
  echo "SPAddr:" >> $file
  echo "  - $sp0_opaddr" >> $file
  echo "  - $sp1_opaddr" >> $file
  echo "  - $sp2_opaddr" >> $file
  echo "  - $sp3_opaddr" >> $file
  echo "  - $sp4_opaddr" >> $file
  echo "  - $sp5_opaddr" >> $file
  echo "  - $sp6_opaddr" >> $file

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
  echo "succeed to stop storage providers"
}

#############################################
# drop databases and recreate new databases #
#############################################
reset_sql_db() {
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
    cd ${sp_dir}
    source db.info
    hostname=$(echo ${ADDRESS} | cut -d : -f 1)
    port=$(echo ${ADDRESS} | cut -d : -f 2)
    mysql -u ${USER} -h ${hostname} -P ${port} -p${PWD} -e "drop database if exists ${DATABASE}"
    mysql -u ${USER} -h ${hostname} -P ${port} -p${PWD} -e "create database ${DATABASE}"
    echo "succeed to reset sql db in "${sp_dir}
    cd - >/dev/null
  done
}

##########################
# clean piece-store data #
##########################
reset_piece_store() {
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
      cd ${sp_dir}
      rm -rf ./data
      echo "succeed to reset piece store in "${sp_dir}
      cd - >/dev/null
    done
}

##################
# print work dir #
##################
print_work_dir() {
  for sp_dir in ${workspace}/${SP_DEPLOY_DIR}/* ; do
    echo "  "${sp_dir}
  done
}

##############
# rebuild sp #
##############
function rebuild() {
  bash ${workspace}/../../build.sh
  mkdir -p ${workspace}/${SP_DEPLOY_DIR}
  for ((i=0;i<${SP_NUM};i++));do
    mkdir -p ${workspace}/${SP_DEPLOY_DIR}/sp${i}
    cp -rf ${sp_bin} ${workspace}/${SP_DEPLOY_DIR}/sp${i}/${sp_bin_name}${i}
  done
}

#############
# reset sps #
#############
reset_sp() {
  stop_sp
  reset_sql_db
  reset_piece_store
  rebuild
  make_config
}

main() {
  CMD=$1
  case ${CMD} in
  --generate)
    generate_sp_db_info $2 $3 $4 $5
    make_intergation_test_config $2
    ;;
  --reset)
    reset_sp
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