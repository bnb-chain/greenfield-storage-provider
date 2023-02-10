package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonehub"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	configFile = flag.String("config", "./config.toml", "config file path")
	binaryFile = flag.String("binary", "./gnfd-sp", "binary file path")
	cfg        *config.StorageProviderConfig
)

const (
	oneboxDir  = "./test-env"
	destBinary = "onebox-gnfd-sp"
	destConfig = "config.toml"
)

func initConfig() {
	cfg.Service = []string{model.SyncerService}
	cfg.GatewayCfg = gateway.DefaultGatewayConfig
	cfg.UploaderCfg = uploader.DefaultUploaderConfig
	cfg.StoneHubCfg = stonehub.DefaultStoneHubConfig
	cfg.ChallengeCfg = challenge.DefaultChallengeConfig
	cfg.StoneNodeCfg = stonenode.DefaultStoneNodeConfig
	if cfg.SyncerCfg.MetaSqlDBConfig == nil {
		cfg.SyncerCfg.MetaSqlDBConfig = metasql.DefaultMetaSqlDBConfig
	}
	if cfg.SyncerCfg.MetaLevelDBConfig == nil {
		cfg.SyncerCfg.MetaLevelDBConfig = metalevel.DefaultMetaLevelDBConfig
	}
	if cfg.SyncerCfg.PieceConfig == nil {
		cfg.SyncerCfg.PieceConfig = storage.DefaultPieceStoreConfig
	}
}

func runShell(cmdStr string) (string, error) {
	var (
		outMsg bytes.Buffer
		errMsg bytes.Buffer
		err    error
	)
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	cmd.Stdout = &outMsg
	cmd.Stderr = &errMsg
	if err = cmd.Run(); err != nil {
		log.Errorw("failed to run cmd", "cmd", cmd, "error", err, "stderr", errMsg, "stdout", outMsg)
		return outMsg.String(), err
	}
	return outMsg.String(), nil
}

func getProcessNum() (int, error) {
	time.Sleep(10 * time.Second)
	getProcessNumCMD := fmt.Sprintf("ps axu|grep %s | grep -v \"grep\" |wc -l", destBinary)
	processNumStr, err := runShell(getProcessNumCMD)
	if err != nil {
		return 0, err
	}
	processNumStr = strings.ReplaceAll(processNumStr, " ", "")
	processNumStr = strings.ReplaceAll(processNumStr, "\n", "")
	processNum, err := strconv.Atoi(processNumStr)
	if err != nil {
		log.Errorw("failed to get process num", "error", err)
		return 0, err
	}
	return processNum, nil
}

func main() {
	log.Info("begin setup onebox, deploy secondary syncers")

	cfg = config.LoadConfig(*configFile)
	addrList := cfg.StoneNodeCfg.SyncerServiceAddress
	initConfig()

	// clear
	// todo: polish not clear data
	os.RemoveAll(oneboxDir)
	pkillCMD := fmt.Sprintf("kill -9 $(pgrep -f %s)", destBinary)
	runShell(pkillCMD)
	if processNum, err := getProcessNum(); err != nil || processNum != 0 {
		log.Errorw("failed to pkill", "error", err)
		os.Exit(1)
		return
	}

	// setup
	if err := os.Mkdir(oneboxDir, 0777); err != nil {
		log.Errorw("failed to mkdir onebox directory", "error", err)
		os.Exit(1)
		return
	}

	for index, addr := range addrList {
		spDir := oneboxDir + "/sp" + strconv.Itoa(index)
		if err := os.Mkdir(spDir, 0777); err != nil {
			log.Errorw("failed to mkdir onebox sp directory", "error", err)
			os.Exit(1)
			return
		}
		cpCMD := fmt.Sprintf("cp %s %s", *binaryFile, spDir+"/"+destBinary)
		if _, err := runShell(cpCMD); err != nil {
			log.Errorw("failed to cp binary", "error", err)
			os.Exit(1)
			return
		}
		f, err := os.Create(spDir + "/" + destConfig)
		if err != nil {
			log.Errorw("failed to create config", "error", err)
			os.Exit(1)
			return
		}
		cfg.SyncerCfg.Address = addr
		cfg.SyncerCfg.StorageProvider = spDir
		cfg.SyncerCfg.MetaLevelDBConfig.Path = spDir + "/leveldb"
		cfg.SyncerCfg.PieceConfig.Store.BucketURL = spDir + "/piece_store"
		if err = util.TomlSettings.NewEncoder(f).Encode(cfg); err != nil {
			log.Errorw("failed to encode config", "error", err)
			os.Exit(1)
			return
		}
		if err = f.Close(); err != nil {
			log.Errorw("failed to close config", "error", err)
			os.Exit(1)
			return
		}
		nohupStartCMD := fmt.Sprintf("nohup %s/%s -config %s/%s </dev/null >%s/log.txt 2>&1&",
			spDir, destBinary, spDir, destConfig, spDir)
		if _, err = runShell(nohupStartCMD); err != nil {
			log.Errorw("failed to start syncer", "error", err)
			os.Exit(1)
			return
		}
		log.Infow("succeed to setup syncer", "dir", spDir)
	}

	// check
	if processNum, err := getProcessNum(); err != nil || processNum != len(addrList) {
		log.Errorw("failed to setup onebox, syncer maybe down and please check log in ./onebox/sp*/log.txt",
			"expect", len(addrList), "actual", processNum)
		os.Exit(1)
		return
	}
	log.Info("succeed to setup onebox")
}
