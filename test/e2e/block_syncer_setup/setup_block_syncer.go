package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"

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
	configFile = flag.String("config", "./config/config.toml", "config file path")
	binaryFile = flag.String("binary", "./build/gnfd-sp", "binary file path")
	cfg        *config.StorageProviderConfig
)

const (
	oneboxDir  = "./test-env"
	destBinary = "onebox-gnfd-sp"
	destConfig = "config.toml"
)

func initConfig() {
	cfg.Service = []string{model.BlockSyncerService}
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

func main() {
	log.Info("begin setup onebox, deploy secondary syncers")
	cfg = config.LoadConfig(*configFile)
	initConfig()
	os.RemoveAll(oneboxDir)
	pkillCMD := fmt.Sprintf("kill -9 $(pgrep -f %s)", destBinary)
	runShell(pkillCMD)
	// setup
	if err := os.Mkdir(oneboxDir, 0777); err != nil {
		log.Errorw("failed to mkdir onebox directory", "error", err)
		os.Exit(1)
		return
	}
	spDir := oneboxDir + "/sp_block_syncer"
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
	fmt.Println(nohupStartCMD)
	if _, err = runShell(nohupStartCMD); err != nil {
		log.Errorw("failed to start syncer", "error", err)
		os.Exit(1)
		return
	}
	log.Infow("succeed to setup syncer", "dir", spDir)
}
