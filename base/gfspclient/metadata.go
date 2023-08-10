package gfspclient

import (
	"context"

	"github.com/bnb-chain/greenfield/types/resource"
	payment_types "github.com/bnb-chain/greenfield/x/payment/types"
	permission_types "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (s *GfSpClient) GetUserBucketsCount(ctx context.Context, account string, includeRemoved bool, opts ...grpc.DialOption) (int64, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	req := &types.GfSpGetUserBucketsCountRequest{
		AccountId:      account,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetUserBucketsCount(ctx, req)
	if err != nil {
		return 0, ErrRPCUnknown
	}
	return resp.GetCount(), nil
}

func (s *GfSpClient) ListDeletedObjectsByBlockNumberRange(ctx context.Context, spOperatorAddress string, startBlockNumber uint64,
	endBlockNumber uint64, includePrivate bool, opts ...grpc.DialOption) ([]*types.Object, uint64, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, uint64(0), err
	}
	defer conn.Close()
	req := &types.GfSpListDeletedObjectsByBlockNumberRangeRequest{
		StartBlockNumber: int64(startBlockNumber),
		EndBlockNumber:   int64(endBlockNumber),
		IncludePrivate:   includePrivate,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListDeletedObjectsByBlockNumberRange(ctx, req)
	if err != nil {
		return nil, uint64(0), ErrRPCUnknown
	}
	return resp.GetObjects(), uint64(resp.GetEndBlockNumber()), nil
}

func (s *GfSpClient) GetUserBuckets(ctx context.Context, account string, includeRemoved bool, opts ...grpc.DialOption) ([]*types.Bucket, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	req := &types.GfSpGetUserBucketsRequest{
		AccountId:      account,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetUserBuckets(ctx, req)
	if err != nil {
		return nil, ErrRPCUnknown
	}
	return resp.GetBuckets(), nil
}

// ListObjectsByBucketName list objects info by a bucket name
func (s *GfSpClient) ListObjectsByBucketName(ctx context.Context, bucketName string, accountId string, maxKeys uint64,
	startAfter string, continuationToken string, delimiter string, prefix string, includeRemoved bool, opts ...grpc.DialOption) (
	objects []*types.Object, KeyCount uint64, MaxKeys uint64, IsTruncated bool, NextContinuationToken string,
	Name string, Prefix string, Delimiter string, CommonPrefixes []string, ContinuationToken string, err error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, 0, 0, false, "", "", "", "", nil, "", err
	}
	defer conn.Close()

	req := &types.GfSpListObjectsByBucketNameRequest{
		BucketName:        bucketName,
		AccountId:         accountId,
		MaxKeys:           maxKeys,
		StartAfter:        startAfter,
		ContinuationToken: continuationToken,
		Delimiter:         delimiter,
		Prefix:            prefix,
		IncludeRemoved:    includeRemoved,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsByBucketName(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send list objects by bucket name rpc", "error", err)
		return nil, 0, 0, false, "", "", "", "", nil, "", err
	}
	return resp.GetObjects(), resp.GetKeyCount(), resp.GetMaxKeys(), resp.GetIsTruncated(), resp.GetNextContinuationToken(),
		resp.GetName(), resp.GetPrefix(), resp.GetDelimiter(), resp.GetCommonPrefixes(), resp.GetContinuationToken(), nil
}

// GetBucketByBucketName get bucket info by a bucket name
func (s *GfSpClient) GetBucketByBucketName(ctx context.Context, bucketName string, includePrivate bool,
	opts ...grpc.DialOption) (*types.Bucket, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetBucketByBucketNameRequest{
		BucketName:     bucketName,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetBucketByBucketName(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get bucket rpc by bucket name", "error", err)
		return nil, err
	}
	return resp.GetBucket(), nil
}

// GetBucketByBucketID get bucket info by a bucket id
func (s *GfSpClient) GetBucketByBucketID(ctx context.Context, bucketID int64, includePrivate bool,
	opts ...grpc.DialOption) (*types.Bucket, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetBucketByBucketIDRequest{
		BucketId:       bucketID,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetBucketByBucketID(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get bucket by bucket id rpc", "error", err)
		return nil, err
	}
	return resp.GetBucket(), nil
}

// ListExpiredBucketsBySp list buckets that are expired by specific sp
func (s *GfSpClient) ListExpiredBucketsBySp(ctx context.Context, createAt int64, primarySpID uint32,
	limit int64, opts ...grpc.DialOption) ([]*types.Bucket, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpListExpiredBucketsBySpRequest{
		CreateAt:    createAt,
		PrimarySpId: primarySpID,
		Limit:       limit,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListExpiredBucketsBySp(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send list expired buckets by sp rpc", "error", err)
		return nil, err
	}
	return resp.GetBuckets(), nil
}

// GetObjectMeta get object metadata
func (s *GfSpClient) GetObjectMeta(ctx context.Context, objectName string, bucketName string,
	includePrivate bool, opts ...grpc.DialOption) (*types.Object, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetObjectMetaRequest{
		ObjectName:     objectName,
		BucketName:     bucketName,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetObjectMeta(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get object meta rpc", "error", err)
		return nil, err
	}
	return resp.GetObject(), nil
}

// GetPaymentByBucketName get bucket payment info by a bucket name
func (s *GfSpClient) GetPaymentByBucketName(ctx context.Context, bucketName string, includePrivate bool,
	opts ...grpc.DialOption) (*payment_types.StreamRecord, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetPaymentByBucketNameRequest{
		BucketName:     bucketName,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetPaymentByBucketName(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get payment by bucket name rpc", "error", err)
		return nil, err
	}
	return resp.GetStreamRecord(), nil
}

// GetPaymentByBucketID get bucket payment info by a bucket id
func (s *GfSpClient) GetPaymentByBucketID(ctx context.Context, bucketID int64, includePrivate bool,
	opts ...grpc.DialOption) (*payment_types.StreamRecord, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetPaymentByBucketIDRequest{
		BucketId:       bucketID,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetPaymentByBucketID(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get payment by bucket id rpc", "error", err)
		return nil, err
	}
	return resp.GetStreamRecord(), nil
}

// VerifyPermission Verify the input account’s permission to input items
func (s *GfSpClient) VerifyPermission(ctx context.Context, Operator string, bucketName string, objectName string,
	actionType permission_types.ActionType, opts ...grpc.DialOption) (*permission_types.Effect, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &storage_types.QueryVerifyPermissionRequest{
		Operator:   Operator,
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: actionType,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpVerifyPermission(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send verify permission rpc", "error", err)
		return nil, err
	}
	return &resp.Effect, nil
}

// GetBucketMeta get bucket info along with its related info such as payment
func (s *GfSpClient) GetBucketMeta(ctx context.Context, bucketName string, includePrivate bool,
	opts ...grpc.DialOption) (*types.Bucket, *payment_types.StreamRecord, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetBucketMetaRequest{
		BucketName:     bucketName,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetBucketMeta(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get bucket meta rpc", "error", err)
		return nil, nil, err
	}
	return resp.GetBucket(), resp.GetStreamRecord(), nil
}

// GetEndpointBySpID get endpoint by sp id
func (s *GfSpClient) GetEndpointBySpID(ctx context.Context, spId uint32, opts ...grpc.DialOption) (string, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return "", ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetEndpointBySpIDRequest{
		SpId: spId,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetEndpointBySpID(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get sp by address rpc", "error", err)
		return "", err
	}
	return resp.GetEndpoint(), nil
}

func (s *GfSpClient) GetBucketReadQuota(ctx context.Context, bucket *storage_types.BucketInfo, opts ...grpc.DialOption) (
	uint64, uint64, uint64, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return uint64(0), uint64(0), uint64(0), ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetBucketReadQuotaRequest{
		BucketInfo: bucket,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetBucketReadQuota(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get bucket read quota", "error", err)
		return uint64(0), uint64(0), uint64(0), ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return uint64(0), uint64(0), uint64(0), resp.GetErr()
	}
	return resp.GetChargedQuotaSize(), resp.GetSpFreeQuotaSize(), resp.GetConsumedSize(), nil
}

func (s *GfSpClient) ListBucketReadRecord(ctx context.Context, bucket *storage_types.BucketInfo, startTimestampUs,
	endTimestampUs, maxRecordNum int64, opts ...grpc.DialOption) ([]*types.ReadRecord, int64, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, 0, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListBucketReadRecordRequest{
		BucketInfo:       bucket,
		StartTimestampUs: startTimestampUs,
		EndTimestampUs:   endTimestampUs,
		MaxRecordNum:     maxRecordNum,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListBucketReadRecord(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list bucket read record", "error", err)
		return nil, 0, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return nil, 0, resp.GetErr()
	}
	return resp.GetReadRecords(), resp.GetNextStartTimestampUs(), nil
}

func (s *GfSpClient) GetUploadObjectState(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (int32, string, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return 0, "", ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpQueryUploadProgressRequest{
		ObjectId: objectID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpQueryUploadProgress(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get uploading object state", "error", err)
		return 0, "", ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return 0, "", resp.GetErr()
	}
	return int32(resp.GetState()), resp.GetErrDescription(), nil
}

func (s *GfSpClient) GetUploadObjectSegment(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (uint32, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return 0, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpQueryResumableUploadSegmentRequest{
		ObjectId: objectID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpQueryResumableUploadSegment(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get uploading object segment", "error", err)
		return 0, ErrRPCUnknown
	}
	if resp.GetErr() != nil {
		return 0, resp.GetErr()
	}
	return resp.GetSegmentCount(), nil
}

func (s *GfSpClient) GetGroupList(ctx context.Context, name string, prefix string, sourceType string, limit int64,
	offset int64, includeRemoved bool, opts ...grpc.DialOption) ([]*types.Group, int64, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, 0, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetGroupListRequest{
		Name:           name,
		Prefix:         prefix,
		SourceType:     sourceType,
		Limit:          limit,
		Offset:         offset,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetGroupList(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get group list", "error", err)
		return nil, 0, ErrRPCUnknown
	}
	return resp.Groups, resp.Count, nil
}

func (s *GfSpClient) ListBucketsByBucketID(ctx context.Context, bucketIDs []uint64, includeRemoved bool, opts ...grpc.DialOption) (map[uint64]*types.Bucket, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListBucketsByBucketIDRequest{
		BucketIds:      bucketIDs,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListBucketsByBucketID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list buckets by bucket ids", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Buckets, nil
}

func (s *GfSpClient) ListObjectsByObjectID(ctx context.Context, objectIDs []uint64, includeRemoved bool, opts ...grpc.DialOption) (map[uint64]*types.Object, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListObjectsByObjectIDRequest{
		ObjectIds:      objectIDs,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsByObjectID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by object ids", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) ListVirtualGroupFamiliesSpID(ctx context.Context, spID uint32, opts ...grpc.DialOption) ([]*virtual_types.GlobalVirtualGroupFamily, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListVirtualGroupFamiliesBySpIDRequest{SpId: spID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListVirtualGroupFamiliesBySpID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list virtual group families by sp id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.GlobalVirtualGroupFamilies, nil
}

func (s *GfSpClient) GetGlobalVirtualGroupByGvgID(ctx context.Context, gvgID uint32, opts ...grpc.DialOption) (*virtual_types.GlobalVirtualGroup, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetGlobalVirtualGroupByGvgIDRequest{GvgId: gvgID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetGlobalVirtualGroupByGvgID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get global virtual group by gvg id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.GlobalVirtualGroup, nil
}

func (s *GfSpClient) ListObjectsInGVGAndBucket(ctx context.Context, gvgID uint32, bucketID uint64, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.ObjectDetails, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListObjectsInGVGAndBucketRequest{
		GvgId:      gvgID,
		BucketId:   bucketID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsInGVGAndBucket(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by gvg id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) ListObjectsInGVG(ctx context.Context, gvgID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.ObjectDetails, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListObjectsInGVGRequest{
		GvgId:      gvgID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsInGVG(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by gvg id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) ListObjectsByGVGAndBucketForGC(ctx context.Context, gvgID uint32, bucketID uint64, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.ObjectDetails, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListObjectsByGVGAndBucketForGCRequest{
		GvgId:      gvgID,
		BucketId:   bucketID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsByGVGAndBucketForGC(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by gvg and bucket for gc", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) GetVirtualGroupFamily(ctx context.Context, vgfID uint32, opts ...grpc.DialOption) (*virtual_types.GlobalVirtualGroupFamily, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetVirtualGroupFamilyRequest{VgfId: vgfID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetVirtualGroupFamily(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get global virtual group family by vgf id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Vgf, nil
}

func (s *GfSpClient) GetGlobalVirtualGroup(ctx context.Context, bucketID uint64, lvgID uint32, opts ...grpc.DialOption) (*virtual_types.GlobalVirtualGroup, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetGlobalVirtualGroupRequest{
		BucketId: bucketID,
		LvgId:    lvgID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetGlobalVirtualGroup(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get global virtual group by lvg id and bucket id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Gvg, nil
}

func (s *GfSpClient) ListGlobalVirtualGroupsByBucket(ctx context.Context, bucketID uint64, opts ...grpc.DialOption) ([]*virtual_types.GlobalVirtualGroup, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListGlobalVirtualGroupsByBucketRequest{
		BucketId: bucketID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListGlobalVirtualGroupsByBucket(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list global virtual group by bucket id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Groups, nil
}

func (s *GfSpClient) ListGlobalVirtualGroupsBySecondarySP(ctx context.Context, spID uint32, opts ...grpc.DialOption) ([]*virtual_types.GlobalVirtualGroup, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListGlobalVirtualGroupsBySecondarySPRequest{
		SpId: spID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListGlobalVirtualGroupsBySecondarySP(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list global virtual group by secondary sp id", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Groups, nil
}

func (s *GfSpClient) ListMigrateBucketEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) ([]*types.ListMigrateBucketEvents, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListMigrateBucketEventsRequest{
		BlockId: blockID,
		SpId:    spID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListMigrateBucketEvents(ctx, req)
	if err != nil {
		// log.CtxErrorw(ctx, "client failed to list migrate bucket events", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Events, nil
}

func (s *GfSpClient) ListSwapOutEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) ([]*types.ListSwapOutEvents, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListSwapOutEventsRequest{
		BlockId: blockID,
		SpId:    spID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListSwapOutEvents(ctx, req)
	if err != nil {
		// log.CtxErrorw(ctx, "client failed to list migrate swap out events", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Events, nil
}

func (s *GfSpClient) ListSpExitEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) (*types.ListSpExitEvents, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListSpExitEventsRequest{
		BlockId: blockID,
		SpId:    spID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListSpExitEvents(ctx, req)
	if err != nil {
		// log.CtxErrorw(ctx, "client failed to list sp exit events", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Events, nil
}

func (s *GfSpClient) GetObjectByID(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (*storage_types.ObjectInfo, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpListObjectsByObjectIDRequest{
		ObjectIds:      []uint64{objectID},
		IncludeRemoved: false,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsByObjectID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by object ids", "error", err)
		return nil, ErrRPCUnknown
	}
	if len(resp.GetObjects()) == 0 {
		return nil, ErrNoSuchObject
	}
	if _, ok := resp.GetObjects()[objectID]; !ok {
		return nil, ErrNoSuchObject
	}
	if resp.GetObjects()[objectID].GetObjectInfo() == nil {
		return nil, ErrNoSuchObject
	}
	return resp.GetObjects()[objectID].GetObjectInfo(), nil
}

// VerifyPermissionByID Verify the input account’s permission to input source type and resource id
func (s *GfSpClient) VerifyPermissionByID(ctx context.Context, Operator string, resourceType resource.ResourceType, resourceID uint64,
	actionType permission_types.ActionType, opts ...grpc.DialOption) (*permission_types.Effect, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpVerifyPermissionByIDRequest{
		Operator:     Operator,
		ResourceType: resourceType,
		ResourceId:   resourceID,
		ActionType:   actionType,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpVerifyPermissionByID(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send verify permission by id rpc", "error", err)
		return nil, err
	}
	return &resp.Effect, nil
}

func (s *GfSpClient) GetSPInfo(ctx context.Context, operatorAddress string, opts ...grpc.DialOption) (*sptypes.StorageProvider, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetSPInfoRequest{
		OperatorAddress: operatorAddress,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetSPInfo(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get sp info by operator address", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.GetStorageProvider(), nil
}

func (s *GfSpClient) GetStatus(ctx context.Context, opts ...grpc.DialOption) (*types.Status, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetStatusRequest{}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetStatus(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get debug info", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.GetStatus(), nil
}

func (s *GfSpClient) GetUserGroups(ctx context.Context, accountID string, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.GroupMember, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetUserGroupsRequest{
		AccountId:  accountID,
		Limit:      limit,
		StartAfter: startAfter,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetUserGroups(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get user groups", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Groups, nil
}

func (s *GfSpClient) GetGroupMembers(ctx context.Context, groupID uint64, startAfter string, limit uint32, opts ...grpc.DialOption) ([]*types.GroupMember, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRPCUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetGroupMembersRequest{
		GroupId:    groupID,
		Limit:      limit,
		StartAfter: startAfter,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetGroupMembers(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get group members", "error", err)
		return nil, ErrRPCUnknown
	}
	return resp.Groups, nil
}
