package blocksyncer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/test"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type BasicTestSuite struct {
	BlockSyncerE2eBaseSuite
}

func (s *BasicTestSuite) SetupSuite() {
	s.BlockSyncerE2eBaseSuite.SetupSuite()
}

func (s *BasicTestSuite) Test_BlockSyncer() {
	go test.MockChainRPCServer()

	args := []string{"", "-config", "config.toml", "--server", "blocksyncer"}

	go func() {
		if err := App.Run(args); err != nil {
			log.Error(err)
		}
	}()

	time.Sleep(time.Second * 20)

	err := test.Verify(s.T())
	s.Equal(nil, err)
}

func TestBasicTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
