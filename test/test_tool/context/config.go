package context

import (
	"bufio"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/util"
)

const (
	DefaultHistoryFile  = "./.storage_provider_cli_history"
	DefaultStoneHubAddr = "127.0.0.1:5323"
)

type CliConf struct {
	StoneHubAddr string
	HistoryFile  string `yaml:"history_file,omitempty"`
}

func LoadCliConf(file string) (*CliConf, error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	cfg := CliConf{}
	if err = util.TomlSettings.NewDecoder(bufio.NewReader(f)).Decode(&cfg); err != nil {
		return nil, err
	}
	if len(cfg.HistoryFile) == 0 {
		cfg.HistoryFile = DefaultHistoryFile
	}
	return &cfg, nil
}
