package gfspclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mockAddress = "localhost:0"

func TestErrRPCUnknownWithDetail(t *testing.T) {
	err := ErrRPCUnknownWithDetail("mock")
	assert.NotNil(t, err)
}

func TestGfSpClient_Connection(t *testing.T) {
	s := mockBufClient()
	conn, err := s.Connection(context.TODO(), mockAddress)
	assert.Nil(t, err)
	defer conn.Close()
}

func TestGfSpClient_ManagerConnSuccess(t *testing.T) {
	s := mockBufClient()
	conn, err := s.ManagerConn(context.TODO())
	assert.Nil(t, err)
	defer conn.Close()
	assert.NotNil(t, conn)
}

func TestGfSpClient_ManagerConnFailure(t *testing.T) {
	s := mockBufClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.ManagerConn(ctx)
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_ApproverConnSuccess(t *testing.T) {
	s := mockBufClient()
	conn, err := s.ApproverConn(context.TODO())
	assert.Nil(t, err)
	defer conn.Close()
	assert.NotNil(t, conn)
}

func TestGfSpClient_ApproverConnFailure(t *testing.T) {
	s := mockBufClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.ApproverConn(ctx)
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_P2PConnSuccess(t *testing.T) {
	s := mockBufClient()
	conn, err := s.P2PConn(context.TODO())
	assert.Nil(t, err)
	defer conn.Close()
	assert.NotNil(t, conn)
}

func TestGfSpClient_P2PConnFailure(t *testing.T) {
	s := mockBufClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.P2PConn(ctx)
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_SignerConnSuccess(t *testing.T) {
	s := mockBufClient()
	conn, err := s.SignerConn(context.TODO())
	assert.Nil(t, err)
	defer conn.Close()
	assert.NotNil(t, conn)
}

func TestGfSpClient_SignerConnFailure(t *testing.T) {
	s := mockBufClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.SignerConn(ctx)
	assert.Contains(t, err.Error(), context.Canceled.Error())
}

func TestGfSpClient_HTTPClient(t *testing.T) {
	s := mockBufClient()
	result := s.HTTPClient(context.TODO())
	assert.NotNil(t, result)
}

func TestGfSpClient_Close(t *testing.T) {
	s := mockBufClient()
	conn1, err1 := s.ManagerConn(context.TODO())
	assert.Nil(t, err1)
	s.managerConn = conn1
	conn2, err2 := s.ApproverConn(context.TODO())
	assert.Nil(t, err2)
	s.approverConn = conn2
	conn3, err3 := s.P2PConn(context.TODO())
	assert.Nil(t, err3)
	s.p2pConn = conn3
	conn4, err4 := s.SignerConn(context.TODO())
	assert.Nil(t, err4)
	s.signerConn = conn4
	err := s.Close()
	assert.Nil(t, err)
}
