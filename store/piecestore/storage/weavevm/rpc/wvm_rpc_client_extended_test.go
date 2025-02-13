package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	weaveVMtypes "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/types"
)

type MockExtendedRPCCaller struct {
	mock.Mock
}

func (m *MockExtendedRPCCaller) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	// Create a slice with the fixed arguments
	allArgs := []interface{}{ctx, result, method}
	// Append the variadic arguments
	allArgs = append(allArgs, args...)
	// Call the mock with the complete slice, expanded
	ret := m.Called(allArgs...)

	// Set the result based on its type.
	switch v := result.(type) {
	case *string:
		if r := ret.Get(0); r != nil {
			*v = r.(string)
		}
	case *hexutil.Bytes:
		if r := ret.Get(0); r != nil {
			*v = r.(hexutil.Bytes)
		}
	}
	return ret.Error(1)
}

type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, msg)
	return uint64(args.Int(0)), args.Error(1)
}

func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return uint64(args.Int(0)), args.Error(1)
}

func (m *MockEthClient) SendTransaction(ctx context.Context, tx *ethtypes.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethtypes.Receipt, error) {
	args := m.Called(ctx, txHash)
	return nil, args.Error(0)
}

type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) GetAccount(ctx context.Context) (common.Address, error) {
	args := m.Called(ctx)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *MockSigner) SignTransaction(ctx context.Context, signData *weaveVMtypes.SignData) (string, error) {
	args := m.Called(ctx, signData)
	return args.String(0), args.Error(1)
}

func TestSendWvmTransaction_RPCCalls(t *testing.T) {
	testCases := []struct {
		name          string
		to            string
		data          []byte
		tags          [][2]string
		setupMocks    func(*MockExtendedRPCCaller, *MockEthClient, *MockSigner)
		expectRPCCall bool
		expectedError string
	}{
		{
			name: "verify RPC payload format",
			to:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			data: []byte("test data"),
			tags: [][2]string{{"Weavm", "v1.0"}},
			setupMocks: func(mockRPC *MockExtendedRPCCaller, mockEth *MockEthClient, mockSigner *MockSigner) {
				addr := common.HexToAddress("0x1000000000000000000000000000000000000000")
				mockSigner.On("GetAccount", mock.Anything).Return(addr, nil)
				mockEth.On("EstimateGas", mock.Anything, mock.Anything).Return(100000, nil)
				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(1000000000), nil)
				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(1, nil)
				mockSigner.On("SignTransaction", mock.Anything, mock.Anything).Return("0xsignedTx", nil)
				mockRPC.On("CallContext",
					mock.Anything,
					mock.AnythingOfType("*string"),
					"eth_sendWvmTransaction",
					matchWvmTransactionRequest("0xsignedTx", [][2]string{{"Weavm", "v1.0"}}),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*string)
					*result = "0xtxHash"
				}).Return("0xtxHash", nil)
			},
			expectRPCCall: true,
		},
		{
			name: "RPC call failure",
			to:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			data: []byte("test data"),
			tags: [][2]string{{"Weavm", "v1.0"}},
			setupMocks: func(mockRPC *MockExtendedRPCCaller, mockEth *MockEthClient, mockSigner *MockSigner) {
				addr := common.HexToAddress("0x1000000000000000000000000000000000000000")
				mockSigner.On("GetAccount", mock.Anything).Return(addr, nil)
				mockEth.On("EstimateGas", mock.Anything, mock.Anything).Return(100000, nil)
				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(1000000000), nil)
				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(1, nil)
				mockSigner.On("SignTransaction", mock.Anything, mock.Anything).Return("0xsignedTx", nil)
				mockRPC.On("CallContext",
					mock.Anything,
					mock.AnythingOfType("*string"),
					"eth_sendWvmTransaction",
					matchWvmTransactionRequest("0xsignedTx", [][2]string{{"Weavm", "v1.0"}}),
				).Run(func(args mock.Arguments) {
					result := args.Get(1).(*string)
					*result = "0xtxHash"
				}).Return("", errors.New("RPC error"))
			},
			expectRPCCall: true,
			expectedError: "failed to send weave transaction: RPC error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRPC := new(MockExtendedRPCCaller)
			mockEth := new(MockEthClient)
			mockSigner := new(MockSigner)

			tc.setupMocks(mockRPC, mockEth, mockSigner)

			client := &RPCClient{
				client:         mockEth,
				clientExtended: mockRPC,
				signer:         mockSigner,
			}

			txHash, err := client.SendWvmTransaction(context.Background(), tc.to, tc.data, tc.tags)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "0xtxHash", txHash)
			}

			mockRPC.AssertExpectations(t)
			mockEth.AssertExpectations(t)
			mockSigner.AssertExpectations(t)
		})
	}
}

func TestGetWvmTransactionByTag_RPCCalls(t *testing.T) {
	testCases := []struct {
		name          string
		tag           [2]string
		mockResponse  []byte
		setupMocks    func(*MockExtendedRPCCaller)
		expectedError string
		verifyRequest bool
	}{
		{
			name:         "successful retrieval",
			tag:          [2]string{"Weavm", "v1.0"},
			mockResponse: []byte("transaction data"),
			setupMocks: func(mockRPC *MockExtendedRPCCaller) {
				mockRPC.On("CallContext", mock.Anything, mock.AnythingOfType("*hexutil.Bytes"),
					"eth_getWvmTransactionByTag",
					mock.MatchedBy(func(req interface{}) bool {
						if r, ok := req.(*GetWvmTransactionByTagRequest); ok {
							return r.Tag[0] == "Weavm" && r.Tag[1] == "v1.0"
						}
						return false
					})).
					Run(func(args mock.Arguments) {
						result := args.Get(1).(*hexutil.Bytes)
						*result = hexutil.Bytes([]byte("transaction data"))
					}).
					Return(hexutil.Bytes([]byte("transaction data")), nil)
			},
			verifyRequest: true,
		},
		{
			name: "RPC call failure",
			tag:  [2]string{"Weavm", "v1.0"},
			setupMocks: func(mockRPC *MockExtendedRPCCaller) {
				mockRPC.On("CallContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("RPC error"))
			},
			expectedError: "failed to get transaction by tag: RPC error",
		},
		{
			name: "empty response",
			tag:  [2]string{"Weavm", "v1.0"},
			setupMocks: func(mockRPC *MockExtendedRPCCaller) {
				mockRPC.On("CallContext", mock.Anything, mock.AnythingOfType("*hexutil.Bytes"),
					"eth_getWvmTransactionByTag", mock.Anything).
					Run(func(args mock.Arguments) {
						result := args.Get(1).(*hexutil.Bytes)
						*result = hexutil.Bytes{}
					}).
					Return(nil, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRPC := new(MockExtendedRPCCaller)
			tc.setupMocks(mockRPC)

			client := &RPCClient{
				clientExtended: mockRPC,
			}

			result, err := client.GetWvmTransactionByTag(context.Background(), tc.tag)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				if tc.mockResponse != nil {
					assert.Equal(t, tc.mockResponse, result)
				}
			}

			if tc.verifyRequest {
				mockRPC.AssertCalled(t, "CallContext", mock.Anything, mock.Anything,
					"eth_getWvmTransactionByTag", mock.Anything)
			}
			mockRPC.AssertExpectations(t)
		})
	}
}

func TestRPCRequestStructs(t *testing.T) {
	t.Run("WvmTransactionRequest serialization", func(t *testing.T) {
		req := WvmTransactionRequest{
			Tx:   "0xsignedTx",
			Tags: [][2]string{{"Weavm", "v1.0"}},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		expected := `{"tx":"0xsignedTx","tags":[["Weavm","v1.0"]]}`
		assert.JSONEq(t, expected, string(data))

		var decoded WvmTransactionRequest
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, req, decoded)
	})

	t.Run("GetWvmTransactionByTagRequest serialization", func(t *testing.T) {
		req := GetWvmTransactionByTagRequest{
			Tag: [2]string{"Weavm", "v1.0"},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		expected := `{"tag":["Weavm","v1.0"]}`
		assert.JSONEq(t, expected, string(data))

		var decoded GetWvmTransactionByTagRequest
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, req, decoded)
	})
}

func matchWvmTransactionRequest(expectedTx string, expectedTags [][2]string) interface{} {
	return mock.MatchedBy(func(req interface{}) bool {
		txReq, ok := req.(*WvmTransactionRequest)
		if !ok {
			return false
		}
		if txReq.Tx != expectedTx {
			return false
		}
		if len(txReq.Tags) != len(expectedTags) {
			return false
		}
		for i, tag := range txReq.Tags {
			if tag[0] != expectedTags[i][0] || tag[1] != expectedTags[i][1] {
				return false
			}
		}
		return true
	})
}
