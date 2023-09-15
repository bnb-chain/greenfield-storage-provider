package gfspclient

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	mockBufNet      = "bufNet"
	bufSize         = 1024 * 1024
	emptyString     = ""
	mockBucketName1 = "mockBucketName1"
	mockBucketName2 = "mockBucketName2"
	mockBucketName3 = "mockBucketName3"
	mockObjectName1 = "mockObjectName1"
	mockObjectName2 = "mockObjectName2"
	mockObjectName3 = "mockObjectName3"
	mockTxHash      = "txHash"
)

var (
	lis           *bufconn.Listener
	mockRPCErr    = errors.New("mock rpc error")
	mockSignature = []byte("mockSignature")
)

func TestMain(m *testing.M) {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	gfspserver.RegisterGfSpApprovalServiceServer(s, &mockApproverServer{})
	gfspserver.RegisterGfSpAuthenticationServiceServer(s, &mockAuthenticatorServer{})
	gfspserver.RegisterGfSpDownloadServiceServer(s, &mockDownloaderServer{})
	gfspserver.RegisterGfSpManageServiceServer(s, &mockManagerServer{})
	gfspserver.RegisterGfSpP2PServiceServer(s, &mockP2PServer{})
	gfspserver.RegisterGfSpQueryTaskServiceServer(s, &mockQueryServer{})
	gfspserver.RegisterGfSpReceiveServiceServer(s, &mockReceiverServer{})
	gfspserver.RegisterGfSpSignServiceServer(s, &mockSignerServer{})
	gfspserver.RegisterGfSpUploadServiceServer(s, &mockUploaderServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	exitCode := m.Run()
	os.Exit(exitCode)
}

func mockBufClient() *GfSpClient {
	return NewGfSpClient(mockBufNet, mockBufNet, mockBufNet, mockBufNet, mockBufNet, mockBufNet, mockBufNet,
		mockBufNet, mockBufNet, true)
}

// setup used for approver, manager, p2p and signer
func setup(t *testing.T, ctx context.Context) *GfSpClient {
	s := mockBufClient()
	conn, err := grpc.DialContext(ctx, mockBufNet, grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	s.approverConn = conn
	s.managerConn = conn
	s.p2pConn = conn
	s.signerConn = conn
	return s
}

func bufDialer(ctx context.Context, address string) (net.Conn, error) {
	return lis.Dial()
}

type mockApproverServer struct{}

func (*mockApproverServer) GfSpAskApproval(ctx context.Context, req *gfspserver.GfSpAskApprovalRequest) (
	*gfspserver.GfSpAskApprovalResponse, error) {
	switch req.Request.(type) {
	case *gfspserver.GfSpAskApprovalRequest_CreateBucketApprovalTask:
		if req.GetCreateBucketApprovalTask().GetCreateBucketInfo().GetBucketName() == mockBucketName1 {
			return nil, mockRPCErr
		} else if req.GetCreateBucketApprovalTask().GetCreateBucketInfo().GetBucketName() == mockBucketName2 {
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpAskApprovalResponse{
				Allowed: true,
				Response: &gfspserver.GfSpAskApprovalResponse_CreateBucketApprovalTask{
					CreateBucketApprovalTask: &gfsptask.GfSpCreateBucketApprovalTask{
						Task:             &gfsptask.GfSpTask{},
						CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName3},
					}},
			}, nil
		}
	case *gfspserver.GfSpAskApprovalRequest_MigrateBucketApprovalTask:
		if req.GetMigrateBucketApprovalTask().GetMigrateBucketInfo().GetBucketName() == mockBucketName1 {
			return nil, mockRPCErr
		} else if req.GetMigrateBucketApprovalTask().GetMigrateBucketInfo().GetBucketName() == mockBucketName2 {
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpAskApprovalResponse{
				Allowed: true,
				Response: &gfspserver.GfSpAskApprovalResponse_MigrateBucketApprovalTask{
					MigrateBucketApprovalTask: &gfsptask.GfSpMigrateBucketApprovalTask{
						Task:              &gfsptask.GfSpTask{},
						MigrateBucketInfo: &storagetypes.MsgMigrateBucket{BucketName: mockBucketName3},
					}},
			}, nil
		}
	case *gfspserver.GfSpAskApprovalRequest_CreateObjectApprovalTask:
		if req.GetCreateObjectApprovalTask().GetCreateObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetCreateObjectApprovalTask().GetCreateObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpAskApprovalResponse{Err: ErrNoSuchObject}, nil
		} else {
			return &gfspserver.GfSpAskApprovalResponse{
				Allowed: true,
				Response: &gfspserver.GfSpAskApprovalResponse_CreateObjectApprovalTask{
					CreateObjectApprovalTask: &gfsptask.GfSpCreateObjectApprovalTask{
						Task: &gfsptask.GfSpTask{},
						CreateObjectInfo: &storagetypes.MsgCreateObject{
							BucketName: mockBucketName3,
							ObjectName: mockObjectName3,
						},
					}},
			}, nil
		}
	default:
		return nil, ErrTypeMismatch
	}
}

type mockAuthenticatorServer struct{}

func (mockAuthenticatorServer) GfSpVerifyAuthentication(ctx context.Context, req *gfspserver.GfSpAuthenticationRequest) (
	*gfspserver.GfSpAuthenticationResponse, error) {
	if req.GetAuthType() == int32(coremodule.AuthOpAskMigrateBucketApproval) {
		return nil, mockRPCErr
	} else if req.GetAuthType() == int32(coremodule.AuthOpAskCreateObjectApproval) {
		return &gfspserver.GfSpAuthenticationResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpAuthenticationResponse{Allowed: true}, nil
	}
}

func (mockAuthenticatorServer) GetAuthNonce(ctx context.Context, req *gfspserver.GetAuthNonceRequest) (
	*gfspserver.GetAuthNonceResponse, error) {
	if req.GetAccountId() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetAccountId() == mockObjectName2 {
		return &gfspserver.GetAuthNonceResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GetAuthNonceResponse{CurrentNonce: 1}, nil
	}
}

func (mockAuthenticatorServer) UpdateUserPublicKey(ctx context.Context, req *gfspserver.UpdateUserPublicKeyRequest) (
	*gfspserver.UpdateUserPublicKeyResponse, error) {
	if req.GetAccountId() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetAccountId() == mockObjectName2 {
		return &gfspserver.UpdateUserPublicKeyResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.UpdateUserPublicKeyResponse{Result: true}, nil
	}
}

func (mockAuthenticatorServer) VerifyGNFD1EddsaSignature(ctx context.Context, req *gfspserver.VerifyGNFD1EddsaSignatureRequest) (
	*gfspserver.VerifyGNFD1EddsaSignatureResponse, error) {
	if req.GetAccountId() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetAccountId() == mockObjectName2 {
		return &gfspserver.VerifyGNFD1EddsaSignatureResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.VerifyGNFD1EddsaSignatureResponse{Result: true}, nil
	}
}

type mockDownloaderServer struct{}

func (mockDownloaderServer) GfSpDownloadObject(ctx context.Context, req *gfspserver.GfSpDownloadObjectRequest) (
	*gfspserver.GfSpDownloadObjectResponse, error) {
	if req.GetDownloadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetDownloadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
		return &gfspserver.GfSpDownloadObjectResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpDownloadObjectResponse{Data: []byte(mockBufNet)}, nil
	}
}

func (mockDownloaderServer) GfSpDownloadPiece(ctx context.Context, req *gfspserver.GfSpDownloadPieceRequest) (
	*gfspserver.GfSpDownloadPieceResponse, error) {
	if req.GetDownloadPieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetDownloadPieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
		return &gfspserver.GfSpDownloadPieceResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpDownloadPieceResponse{Data: []byte(mockBufNet)}, nil
	}
}

func (mockDownloaderServer) GfSpGetChallengeInfo(ctx context.Context, req *gfspserver.GfSpGetChallengeInfoRequest) (
	*gfspserver.GfSpGetChallengeInfoResponse, error) {
	if req.GetChallengePieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetChallengePieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
		return &gfspserver.GfSpGetChallengeInfoResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpGetChallengeInfoResponse{Data: []byte(mockBufNet)}, nil
	}
}

func (mockDownloaderServer) GfSpReimburseQuota(ctx context.Context, req *gfspserver.GfSpReimburseQuotaRequest) (
	*gfspserver.GfSpReimburseQuotaResponse, error) {
	return &gfspserver.GfSpReimburseQuotaResponse{Err: ErrExceptionsStream}, nil
}

type mockManagerServer struct{}

func (mockManagerServer) GfSpBeginTask(ctx context.Context, req *gfspserver.GfSpBeginTaskRequest) (
	*gfspserver.GfSpBeginTaskResponse, error) {
	switch req.Request.(type) {
	case *gfspserver.GfSpBeginTaskRequest_UploadObjectTask:
		if req.GetUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpBeginTaskResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpBeginTaskResponse{}, nil
		}
	case *gfspserver.GfSpBeginTaskRequest_ResumableUploadObjectTask:
		if req.GetResumableUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetResumableUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpBeginTaskResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpBeginTaskResponse{}, nil
		}
	default:
		return nil, ErrTypeMismatch
	}
}

func (mockManagerServer) GfSpAskTask(ctx context.Context, req *gfspserver.GfSpAskTaskRequest) (
	*gfspserver.GfSpAskTaskResponse, error) {
	switch req.GetNodeLimit().GetMemoryLimit() {
	case -2:
		return nil, mockRPCErr
	case -1:
		return &gfspserver.GfSpAskTaskResponse{Err: ErrExceptionsStream}, nil
	case 0:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_ReplicatePieceTask{
			ReplicatePieceTask: &gfsptask.GfSpReplicatePieceTask{},
		}}, nil
	case 1:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_SealObjectTask{
			SealObjectTask: &gfsptask.GfSpSealObjectTask{},
		}}, nil
	case 2:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_ReceivePieceTask{
			ReceivePieceTask: &gfsptask.GfSpReceivePieceTask{},
		}}, nil
	case 3:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_GcObjectTask{
			GcObjectTask: &gfsptask.GfSpGCObjectTask{},
		}}, nil
	case 4:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_GcZombiePieceTask{
			GcZombiePieceTask: &gfsptask.GfSpGCZombiePieceTask{},
		}}, nil
	case 5:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_GcMetaTask{
			GcMetaTask: &gfsptask.GfSpGCMetaTask{},
		}}, nil
	case 6:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_RecoverPieceTask{
			RecoverPieceTask: &gfsptask.GfSpRecoverPieceTask{},
		}}, nil
	case 7:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_MigrateGvgTask{
			MigrateGvgTask: &gfsptask.GfSpMigrateGVGTask{},
		}}, nil
	case 8:
		return &gfspserver.GfSpAskTaskResponse{Response: &gfspserver.GfSpAskTaskResponse_MigrateGvgTask{}}, nil
	default:
		return nil, ErrTypeMismatch
	}
}

func (mockManagerServer) GfSpReportTask(ctx context.Context, req *gfspserver.GfSpReportTaskRequest) (
	*gfspserver.GfSpReportTaskResponse, error) {
	if req.GetUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
		return &gfspserver.GfSpReportTaskResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpReportTaskResponse{}, nil
	}
}

func (mockManagerServer) GfSpPickVirtualGroupFamily(ctx context.Context, req *gfspserver.GfSpPickVirtualGroupFamilyRequest) (
	*gfspserver.GfSpPickVirtualGroupFamilyResponse, error) {
	if req.GetCreateBucketApprovalTask().GetCreateBucketInfo().GetBucketName() == mockBucketName1 {
		return nil, mockRPCErr
	} else if req.GetCreateBucketApprovalTask().GetCreateBucketInfo().GetBucketName() == mockBucketName2 {
		return &gfspserver.GfSpPickVirtualGroupFamilyResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpPickVirtualGroupFamilyResponse{VgfId: 1}, nil
	}
}

func (mockManagerServer) GfSpNotifyMigrateSwapOut(ctx context.Context, req *gfspserver.GfSpNotifyMigrateSwapOutRequest) (
	*gfspserver.GfSpNotifyMigrateSwapOutResponse, error) {
	if req.GetSwapOut().GlobalVirtualGroupFamilyId == 0 {
		return nil, mockRPCErr
	} else if req.GetSwapOut().GlobalVirtualGroupFamilyId == 1 {
		return &gfspserver.GfSpNotifyMigrateSwapOutResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpNotifyMigrateSwapOutResponse{}, nil
	}
}

func (s mockManagerServer) GfSpQueryTasksStats(ctx context.Context, req *gfspserver.GfSpQueryTasksStatsRequest) (*gfspserver.GfSpQueryTasksStatsResponse, error) {
	if req == nil {
		return nil, mockRPCErr
	}
	return &gfspserver.GfSpQueryTasksStatsResponse{
		Stats: &gfspserver.TasksStats{},
	}, nil
}

type mockP2PServer struct{}

func (mockP2PServer) GfSpAskSecondaryReplicatePieceApproval(ctx context.Context, req *gfspserver.GfSpAskSecondaryReplicatePieceApprovalRequest) (
	*gfspserver.GfSpAskSecondaryReplicatePieceApprovalResponse, error) {
	if req.GetMin() == -2 {
		return nil, mockRPCErr
	} else if req.GetMin() == -1 {
		return &gfspserver.GfSpAskSecondaryReplicatePieceApprovalResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpAskSecondaryReplicatePieceApprovalResponse{
			ApprovedTasks: []*gfsptask.GfSpReplicatePieceApprovalTask{
				{
					Task:       &gfsptask.GfSpTask{},
					ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName3},
				},
			},
		}, nil
	}
}

func (mockP2PServer) GfSpQueryP2PBootstrap(ctx context.Context, req *gfspserver.GfSpQueryP2PNodeRequest) (*gfspserver.GfSpQueryP2PNodeResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("failed to get metadata")
	}
	if v, ok := md["bufnet"]; ok {
		for _, j := range v {
			if j == mockObjectName1 {
				return nil, mockRPCErr
			} else if j == mockObjectName2 {
				return &gfspserver.GfSpQueryP2PNodeResponse{Err: ErrExceptionsStream}, nil
			} else {
				return &gfspserver.GfSpQueryP2PNodeResponse{Nodes: []string{mockObjectName3}}, nil
			}
		}
	}
	return nil, nil
}

type mockQueryServer struct{}

func (mockQueryServer) GfSpQueryTasks(ctx context.Context, req *gfspserver.GfSpQueryTasksRequest) (
	*gfspserver.GfSpQueryTasksResponse, error) {
	if req.GetTaskSubKey() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetTaskSubKey() == mockObjectName2 {
		return &gfspserver.GfSpQueryTasksResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpQueryTasksResponse{TaskInfo: []string{mockBufNet}}, nil
	}
}

func (mockQueryServer) GfSpQueryBucketMigrate(ctx context.Context, req *gfspserver.GfSpQueryBucketMigrateRequest) (
	*gfspserver.GfSpQueryBucketMigrateResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("failed to get metadata")
	}
	if v, ok := md["bufnet"]; ok {
		for _, j := range v {
			if j == mockObjectName1 {
				return nil, mockRPCErr
			} else if j == mockObjectName2 {
				return &gfspserver.GfSpQueryBucketMigrateResponse{Err: ErrExceptionsStream}, nil
			} else {
				return &gfspserver.GfSpQueryBucketMigrateResponse{SelfSpId: 1}, nil
			}
		}
	}
	return nil, nil
}

func (mockQueryServer) GfSpQuerySpExit(ctx context.Context, req *gfspserver.GfSpQuerySpExitRequest) (
	*gfspserver.GfSpQuerySpExitResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("failed to get metadata")
	}
	if v, ok := md["bufnet"]; ok {
		for _, j := range v {
			if j == mockObjectName1 {
				return nil, mockRPCErr
			} else if j == mockObjectName2 {
				return &gfspserver.GfSpQuerySpExitResponse{Err: ErrExceptionsStream}, nil
			} else {
				return &gfspserver.GfSpQuerySpExitResponse{SelfSpId: 1}, nil
			}
		}
	}
	return nil, nil
}

func (mockManagerServer) GfSpQuerySPByOperatorAddress(ctx context.Context, req *gfspserver.GfSpQuerySPByOperatorAddressRequest) (
	*gfspserver.GfSpQuerySPByOperatorAddressResponse, error) {
	return nil, nil
}

type mockReceiverServer struct{}

func (mockReceiverServer) GfSpReplicatePiece(ctx context.Context, req *gfspserver.GfSpReplicatePieceRequest) (
	*gfspserver.GfSpReplicatePieceResponse, error) {
	if req.GetReceivePieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetReceivePieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
		return &gfspserver.GfSpReplicatePieceResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpReplicatePieceResponse{}, nil
	}
}

func (mockReceiverServer) GfSpDoneReplicatePiece(ctx context.Context, req *gfspserver.GfSpDoneReplicatePieceRequest) (
	*gfspserver.GfSpDoneReplicatePieceResponse, error) {
	if req.GetReceivePieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
		return nil, mockRPCErr
	} else if req.GetReceivePieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
		return &gfspserver.GfSpDoneReplicatePieceResponse{Err: ErrExceptionsStream}, nil
	} else {
		return &gfspserver.GfSpDoneReplicatePieceResponse{Signature: mockSignature}, nil
	}
}

type mockSignerServer struct{}

func (mockSignerServer) GfSpSign(ctx context.Context, req *gfspserver.GfSpSignRequest) (*gfspserver.GfSpSignResponse, error) {
	switch req.Request.(type) {
	case *gfspserver.GfSpSignRequest_CreateBucketInfo:
		if req.GetCreateBucketInfo().GetBucketName() == mockBucketName1 {
			return nil, mockRPCErr
		} else if req.GetCreateBucketInfo().GetBucketName() == mockBucketName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_MigrateBucketInfo:
		if req.GetMigrateBucketInfo().GetBucketName() == mockBucketName1 {
			return nil, mockRPCErr
		} else if req.GetMigrateBucketInfo().GetBucketName() == mockBucketName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_CreateObjectInfo:
		if req.GetCreateObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetCreateObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_SealObjectInfo:
		if req.GetSealObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetSealObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_SpStoragePrice:
		if req.GetSpStoragePrice().GetFreeReadQuota() == 0 {
			return nil, mockRPCErr
		} else if req.GetSpStoragePrice().GetFreeReadQuota() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup:
		if req.GetCreateGlobalVirtualGroup().GetVirtualGroupFamilyId() == 0 {
			return nil, mockRPCErr
		} else if req.GetCreateGlobalVirtualGroup().GetVirtualGroupFamilyId() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_RejectObjectInfo:
		if req.GetRejectObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetRejectObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_DiscontinueBucketInfo:
		if req.GetDiscontinueBucketInfo().GetBucketName() == mockBucketName1 {
			return nil, mockRPCErr
		} else if req.GetDiscontinueBucketInfo().GetBucketName() == mockBucketName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask:
		if req.GetGfspReplicatePieceApprovalTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetGfspReplicatePieceApprovalTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_SignSecondarySealBls:
		if req.GetSignSecondarySealBls().GetObjectId() == 0 {
			return nil, mockRPCErr
		} else if req.GetSignSecondarySealBls().GetObjectId() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_GfspReceivePieceTask:
		if req.GetGfspReceivePieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetGfspReceivePieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_GfspRecoverPieceTask:
		if req.GetGfspRecoverPieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetGfspRecoverPieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_PingMsg:
		if req.GetPingMsg().GetSpOperatorAddress() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetPingMsg().GetSpOperatorAddress() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_PongMsg:
		if req.GetPongMsg().GetSpOperatorAddress() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetPongMsg().GetSpOperatorAddress() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_GfspMigratePieceTask:
		if req.GetGfspMigratePieceTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetGfspMigratePieceTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_CompleteMigrateBucket:
		if req.GetCompleteMigrateBucket().GetBucketName() == mockBucketName1 {
			return nil, mockRPCErr
		} else if req.GetCompleteMigrateBucket().GetBucketName() == mockBucketName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_SignSecondarySpMigrationBucket:
		if req.GetSignSecondarySpMigrationBucket().GetDstPrimarySpId() == 0 {
			return nil, mockRPCErr
		} else if req.GetSignSecondarySpMigrationBucket().GetDstPrimarySpId() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_SignSwapOut:
		if req.GetSignSwapOut().GetGlobalVirtualGroupFamilyId() == 0 {
			return nil, mockRPCErr
		} else if req.GetSignSwapOut().GetGlobalVirtualGroupFamilyId() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{Signature: mockSignature}, nil
		}
	case *gfspserver.GfSpSignRequest_SwapOut:
		if req.GetSwapOut().GetGlobalVirtualGroupFamilyId() == 0 {
			return nil, mockRPCErr
		} else if req.GetSwapOut().GetGlobalVirtualGroupFamilyId() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_CompleteSwapOut:
		if req.GetCompleteSwapOut().GetGlobalVirtualGroupFamilyId() == 0 {
			return nil, mockRPCErr
		} else if req.GetCompleteSwapOut().GetGlobalVirtualGroupFamilyId() == 1 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_SpExit:
		if req.GetSpExit().GetStorageProvider() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetSpExit().GetStorageProvider() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	case *gfspserver.GfSpSignRequest_CompleteSpExit:
		if req.GetCompleteSpExit().GetStorageProvider() == mockObjectName1 {
			return nil, mockRPCErr
		} else if req.GetCompleteSpExit().GetStorageProvider() == mockObjectName2 {
			return &gfspserver.GfSpSignResponse{Err: ErrExceptionsStream}, nil
		} else {
			return &gfspserver.GfSpSignResponse{TxHash: mockTxHash}, nil
		}
	default:
		return nil, ErrTypeMismatch
	}
}

type mockUploaderServer struct{}

func (mockUploaderServer) GfSpUploadObject(stream gfspserver.GfSpUploadService_GfSpUploadObjectServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&gfspserver.GfSpUploadObjectResponse{})
		}
		if err != nil {
			return err
		}
		if req.GetUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return mockRPCErr
		} else if req.GetUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return stream.SendAndClose(&gfspserver.GfSpUploadObjectResponse{Err: ErrNoSuchObject})
		}
	}
}

func (mockUploaderServer) GfSpResumableUploadObject(stream gfspserver.GfSpUploadService_GfSpResumableUploadObjectServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&gfspserver.GfSpResumableUploadObjectResponse{})
		}
		if err != nil {
			return err
		}
		if req.GetResumableUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName1 {
			return mockRPCErr
		} else if req.GetResumableUploadObjectTask().GetObjectInfo().GetObjectName() == mockObjectName2 {
			return stream.SendAndClose(&gfspserver.GfSpResumableUploadObjectResponse{Err: ErrNoSuchObject})
		}
	}
}
