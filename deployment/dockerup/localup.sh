#!/usr/bin/env bash

basedir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
workspace=${basedir}
source "${workspace}/env.info"
bin_name=mechain-sp
bin="$(realpath "${workspace}/../../build/${bin_name}")"

#########################
# the command line help #
#########################
function display_help() {
  echo "Usage: $0 [option...] {help|generate|reset|start|stop|print}" >&2
  echo
  echo "   --help           display help info"
  echo "   --generate       generate sp.info and db.info that accepts four args: the first arg is json file path, the second arg is db username, the third arg is db password and the fourth arg is db address"
  echo "   --reset          reset env"
  echo "   --start          start storage providers"
  echo "   --stop           stop storage providers"
  echo "   --clean          clean local sp env"
  echo "   --rebuild        rebuild sp code"
  echo "   --print          print sp local env work directory"
  echo
  exit 0
}

################################
# generate sp.info and db.info #
################################
function generate_sp_db_info() {
  if [ $# != 4 ]; then
    echo "failed to generate sp.info and db.info, please check args by help info"
    exit 1
  fi
  sp_json_file=$1
  db_user=$2
  db_password=$3
  db_address=$4
  for ((i = 0; i < SP_NUM; i++)); do
    mkdir -p "${workspace}/${SP_DEPLOY_DIR}/sp${i}"
    cp -rf "${bin}" "${workspace}/${SP_DEPLOY_DIR}/sp${i}/${bin_name}${i}"
    cd "${workspace}/${SP_DEPLOY_DIR}/sp${i}" || exit 1
    "${workspace}/${SP_DEPLOY_DIR}/sp${i}/${bin_name}${i}" config.dump

    # generate sp info
    endpoint="0.0.0.0:${SP_START_ENDPOINT_PORT}"
    {
      echo "#!/usr/bin/env bash"
      echo "SP_ENDPOINT=\"${endpoint}\""
    } >sp.info
    op_address=$(jq -r ".sp${i}.OperatorAddress" "${sp_json_file}")
    echo "OPERATOR_ADDRESS=\"${op_address}\"" >>sp.info
    opk=$(jq -r ".sp${i}.OperatorPrivateKey" "${sp_json_file}")
    echo "OPERATOR_PRIVATE_KEY=\"${opk}\"" >>sp.info
    fpk=$(jq -r ".sp${i}.FundingPrivateKey" "${sp_json_file}")
    echo "FUNDING_PRIVATE_KEY=\"${fpk}\"" >>sp.info
    spk=$(jq -r ".sp${i}.SealPrivateKey" "${sp_json_file}")
    echo "SEAL_PRIVATE_KEY=\"${spk}\"" >>sp.info
    apk=$(jq -r ".sp${i}.ApprovalPrivateKey" "${sp_json_file}")
    echo "APPROVAL_PRIVATE_KEY=\"${apk}\"" >>sp.info
    gpk=$(jq -r ".sp${i}.GcPrivateKey" "${sp_json_file}")
    echo "GC_PRIVATE_KEY=\"${gpk}\"" >>sp.info
    bpk=$(jq -r ".sp${i}.BlsPrivateKey" "${sp_json_file}")
    echo "BLS_PRIVATE_KEY=\"${bpk}\"" >>sp.info

    # generate db info
    {
      echo '#!/usr/bin/env bash'
      echo "USER=\"${db_user}\""
      echo "PWD=\"${db_password}\""
      echo "ADDRESS=\"${db_address}\""
      echo "DATABASE=sp_${i}"
    } >db.info
    cd - >/dev/null || exit
  done
  print_work_dir
  echo "succeed to generate sp.info and db.info"
}

#############################################################
# make sp config.toml according to env.info/db.info/sp.info #
#############################################################
function make_config() {
  index=0
  for sp_dir in "${workspace}/${SP_DEPLOY_DIR}"/*; do
    cd "${sp_dir}" || exit 1
    source db.info
    source sp.info
    # app
    sed -i -e "s/GRPCAddress = '.*'/GRPCAddress = '0.0.0.0:${SP_START_PORT}'/g" config.toml

    # db
    sed -i -e "s/User = '.*'/User = '${USER}'/g" config.toml
    sed -i -e "s/Passwd = '.*'/Passwd = '${PWD}'/g" config.toml
    sed -i -e "s/^Address = '.*'/Address = '${ADDRESS}'/g" config.toml
    sed -i -e "s/Database = '.*'/Database = '${DATABASE}'/g" config.toml

    # chain
    sed -i -e "s/ChainID = '.*'/ChainID = '${CHAIN_ID}'/g" config.toml
    sed -i -e "s/ChainAddress = \[.*\]/ChainAddress = \['http:\/\/${CHAIN_HTTP_ENDPOINT}'\]/g" config.toml
    sed -i -e "s/RpcAddress = \[.*\]/RpcAddress = \['http:\/\/${CHAIN_EVM_ENDPOINT}'\]/g" config.toml

    # sp account
    sed -i -e "s/SpOperatorAddress = '.*'/SpOperatorAddress = '${OPERATOR_ADDRESS}'/g" config.toml
    sed -i -e "s/OperatorPrivateKey = '.*'/OperatorPrivateKey = '${OPERATOR_PRIVATE_KEY}'/g" config.toml
    sed -i -e "s/FundingPrivateKey = '.*'/FundingPrivateKey = '${FUNDING_PRIVATE_KEY}'/g" config.toml
    sed -i -e "s/SealPrivateKey = '.*'/SealPrivateKey = '${SEAL_PRIVATE_KEY}'/g" config.toml
    sed -i -e "s/ApprovalPrivateKey = '.*'/ApprovalPrivateKey = '${APPROVAL_PRIVATE_KEY}'/g" config.toml
    sed -i -e "s/GcPrivateKey = '.*'/GcPrivateKey = '${GC_PRIVATE_KEY}'/g" config.toml
    sed -i -e "s/BlsPrivateKey = '.*'/BlsPrivateKey = '${BLS_PRIVATE_KEY}'/g" config.toml

    # gateway
    sed -i -e "s/DomainName = '.*'/DomainName = 'gnfd.test-sp.com'/g" config.toml
    sed -i -e "s/^HTTPAddress = '.*'/HTTPAddress = '${SP_ENDPOINT}'/g" config.toml

    # metadata
    sed -i -e "s/IsMasterDB = .*/IsMasterDB = true/g" config.toml
    sed -i -e "s/BsDBSwitchCheckIntervalSec = .*/BsDBSwitchCheckIntervalSec = 30/g" config.toml

    # p2p
    sed -i -e "s/P2PAddress = '.*'/P2PAddress = '0.0.0.0:9633'/g" config.toml
    if [ ${index} -eq 0 ]; then
      sed -i -e "s/P2PPrivateKey = '.*'/P2PPrivateKey = '${SP0_P2P_PRIVATE_KEY}'/g" config.toml
    else
      sed -i -e "s/Bootstrap = \[\]/Bootstrap = \[\'16Uiu2HAmG4KTyFsK71BVwjY4z6WwcNBVb6vAiuuL9ASWdqiTzNZH@0.0.0.0:9633\'\]/g" config.toml
    fi

    sed -i -e "s/MaxExecuteNumber = .*/MaxExecuteNumber = 1/g" config.toml

    # metrics and pprof
    #sed -i -e "s/DisableMetrics = false/DisableMetrics = true/" config.toml
    #sed -i -e "s/DisablePProf = false/DisablePProf = true/" config.toml
    #sed -i -e "s/DisableProbe = false/DisableProbe = true/" config.toml
    metrics_address="0.0.0.0:"$((SP_START_PORT + 367))
    sed -i -e "s/MetricsHTTPAddress = '.*'/MetricsHTTPAddress = '${metrics_address}'/g" config.toml
    pprof_address="0.0.0.0:"$((SP_START_PORT + 368))
    sed -i -e "s/PProfHTTPAddress = '.*'/PProfHTTPAddress = '${pprof_address}'/g" config.toml
    probe_address="0.0.0.0:"$((SP_START_PORT + 369))
    sed -i -e "s/ProbeHTTPAddress = '.*'/ProbeHTTPAddress = '${probe_address}'/g" config.toml

    # blocksyncer
    sed -i -e "s/Modules = \[\]/Modules = \[\'epoch\',\'bucket\',\'object\',\'payment\',\'group\',\'permission\',\'storage_provider\'\,\'prefix_tree\'\,\'virtual_group\'\,\'sp_exit_events\'\,\'object_id_map\'\,\'general\'\]/g" config.toml
    WORKERS=10
    sed -i -e "s/Workers = 0/Workers = ${WORKERS}/g" config.toml
    sed -i -e "s/BsDBWriteAddress = '.*'/BsDBWriteAddress = '${ADDRESS}'/g" config.toml

    # manager
    sed -i -e "s/SubscribeSPExitEventIntervalMillisecond = .*/SubscribeSPExitEventIntervalMillisecond = 100/g" config.toml
    sed -i -e "s/SubscribeSwapOutExitEventIntervalMillisecond = .*/SubscribeSwapOutExitEventIntervalMillisecond = 100/g" config.toml
    sed -i -e "s/SubscribeBucketMigrateEventIntervalMillisecond = .*/SubscribeBucketMigrateEventIntervalMillisecond = 20/g" config.toml
    sed -i -e "s/GVGPreferSPList = \[\]/GVGPreferSPList = \[1,2,3,4,5,6,7,8\]/g" config.toml
    sed -i -e "s/EnableGCZombie = .*/EnableGCZombie = true/g" config.toml
    sed -i -e "s/EnableGCMeta = .*/EnableGCMeta = true/g" config.toml
    sed -i -e "s/GCMetaTimeInterval = .*/GCMetaTimeInterval = 3/g" config.toml
    sed -i -e "s/GCZombiePieceTimeInterval = .*/GCZombiePieceTimeInterval = 3/g" config.toml
    sed -i -e "s/GCZombieSafeObjectIDDistance = .*/GCZombieSafeObjectIDDistance = 1/g" config.toml
    sed -i -e "s/GCZombiePieceObjectIDInterval = .*/GCZombiePieceObjectIDInterval = 5/g" config.toml
    sed -i -e "s/EnableTaskRetryScheduler = .*/EnableTaskRetryScheduler = true/g" config.toml
    sed -i -e "s/RejectUnsealThresholdSecond = .*/RejectUnsealThresholdSecond = 600/g" config.toml
    sed -i -e "s/EnableHealthyChecker = .*/EnableHealthyChecker = true/g" config.toml
    sed -i -e "s/EnableGCStaleVersionObject = .*/EnableGCStaleVersionObject = true/g" config.toml
    sed -i -e "s/EnableGCExpiredOffChainAuthKeys = .*/EnableGCExpiredOffChainAuthKeys = true/g" config.toml
    sed -i -e "s/GCExpiredOffChainAuthKeysTimeInterval = .*/GCExpiredOffChainAuthKeysTimeInterval = 86400/g" config.toml
    sed -i -e "s/GasLimit = 0/GasLimit = 180000/g" config.toml
    sed -i -e "s/FeeAmount = 0/FeeAmount = 12000000/g" config.toml

    echo "succeed to generate config.toml in ""${sp_dir}"
    cd - >/dev/null || exit
    index=$(($index + 1))
  done
}

#############
# start sps #
#############
function start_sp() {
  index=0
  for sp_dir in "${workspace}/${SP_DEPLOY_DIR}"/*; do
    cd "${sp_dir}" || exit 1
    nohup ./${bin_name}${index} --config config.toml </dev/null >log.txt 2>&1 &
    echo "succeed to start sp in ""${sp_dir}"
    cd - >/dev/null || exit
    index=$(($index + 1))
  done
  echo "succeed to start storage providers"
}

############
# stop sps #
############
function stop_sp() {
  kill -9 $(LC_ALL=C pgrep -f ${bin_name}) >/dev/null 2>&1
  echo "succeed to stop storage providers"
}

#############################################
# drop databases and recreate new databases #
#############################################
function reset_sql_db() {
  for sp_dir in "${workspace}/${SP_DEPLOY_DIR}"/*; do
    cd "${sp_dir}" || exit 1
    source db.info
    hostname=$(echo "${ADDRESS}" | cut -d : -f 1)
    port=$(echo "${ADDRESS}" | cut -d : -f 2)
    mysql -u "${USER}" -h "${hostname}" -P "${port}" -p"${PWD}" -e "drop database if exists ${DATABASE}"
    mysql -u "${USER}" -h "${hostname}" -P "${port}" -p"${PWD}" -e "create database ${DATABASE}"
    echo "succeed to reset sql db in ""${sp_dir}"
    cd - >/dev/null || exit
  done
}

##########################
# clean piece-store data #
##########################
function reset_piece_store() {
  for sp_dir in "${workspace}/${SP_DEPLOY_DIR}"/*; do
    cd "${sp_dir}" || exit 1
    rm -rf ./data
    echo "succeed to reset piece store in ""${sp_dir}"
    cd - >/dev/null || exit
  done
}

##################
# print work dir #
##################
function print_work_dir() {
  for sp_dir in "${workspace}/${SP_DEPLOY_DIR}"/*; do
    echo "  ""${sp_dir}"
  done
}

######################
# clean local sp env #
######################
function clean_local_sp_env() {
  rm -rf "${workspace:?}/${SP_DEPLOY_DIR:?}"
}

#############
# reset sps #
#############
function reset_sp() {
  stop_sp
  reset_sql_db
  reset_piece_store
  make_config
}

function main() {
  CMD=$1
  case ${CMD} in
  --generate)
    generate_sp_db_info "$2" "$3" "$4" "$5"
    ;;
  --config)
    make_config
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
  --clean)
    clean_local_sp_env
    ;;
  --print)
    print_work_dir
    ;;
  --rebuild)
    rebuild
    ;;
  --help | *)
    display_help
    ;;
  esac
}

main $@
