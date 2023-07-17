package basesuite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/stretchr/testify/suite"
)

var (
	// Endpoint = "gnfd-testnet-fullnode-cosmos-us.nodereal.io:443"
	Endpoint = "http://localhost:26750"
	ChainID  = "greenfield_9000-121"
)

func ParseMnemonicFromFile(fileName string) string {
	fileName = filepath.Clean(fileName)
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	// #nosec
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		if scanner.Text() != "" {
			line = scanner.Text()
		}
	}
	return line
}

type BaseSuite struct {
	suite.Suite
	DefaultAccount *types.Account
	Client         client.Client
	ClientContext  context.Context
}

// ParseValidatorMnemonic read the validator mnemonic from file
func ParseValidatorMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../../greenfield/deployment/localup/.local/validator%d/info", i))
}

func (s *BaseSuite) SetupSuite() {
	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	s.Require().NoError(err)
	s.Client, err = client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
	})
	s.Require().NoError(err)
	s.ClientContext = context.Background()
	s.DefaultAccount = account
}
