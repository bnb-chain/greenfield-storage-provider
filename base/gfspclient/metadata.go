package gfspclient

import (
	"context"

	payment_types "github.com/bnb-chain/greenfield/x/payment/types"
	permission_types "github.com/bnb-chain/greenfield/x/permission/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
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
		return 0, ErrRpcUnknown
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
		return nil, uint64(0), ErrRpcUnknown
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
		return nil, ErrRpcUnknown
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
func (s *GfSpClient) GetBucketByBucketID(ctx context.Context, bucketId int64, includePrivate bool,
	opts ...grpc.DialOption) (*types.GfSpGetBucketByBucketIDResponse, error) {
	conn, err := s.Connection(ctx, s.metadataEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := &types.GfSpGetBucketByBucketIDRequest{
		BucketId:       bucketId,
		IncludePrivate: includePrivate,
	}

	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetBucketByBucketID(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get bucket by bucket id rpc", "error", err)
		return nil, err
	}
	return resp, nil
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

// GetEndpointBySpAddress get endpoint by sp address
func (s *GfSpClient) GetEndpointBySpAddress(ctx context.Context, spAddress string, opts ...grpc.DialOption) (string, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return "", ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetEndpointBySpAddressRequest{
		SpAddress: spAddress,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetEndpointBySpAddress(ctx, req)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "failed to send get sp by address rpc", "error", err)
		return "", err
	}
	return resp.GetEndpoint(), nil
}

func (s *GfSpClient) GetBucketReadQuota(ctx context.Context, bucket *storage_types.BucketInfo, yearMonth string, opts ...grpc.DialOption) (
	uint64, uint64, uint64, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return uint64(0), uint64(0), uint64(0), ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetBucketReadQuotaRequest{
		BucketInfo: bucket,
		YearMonth:  yearMonth,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetBucketReadQuota(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get bucket read quota", "error", err)
		return uint64(0), uint64(0), uint64(0), ErrRpcUnknown
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
		return nil, 0, ErrRpcUnknown
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
		return nil, 0, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return nil, 0, resp.GetErr()
	}
	return resp.GetReadRecords(), resp.GetNextStartTimestampUs(), nil
}

func (s *GfSpClient) GetUploadObjectState(ctx context.Context, objectID uint64, opts ...grpc.DialOption) (int32, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return 0, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpQueryUploadProgressRequest{
		ObjectId: objectID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpQueryUploadProgress(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get uploading object state", "error", err)
		return 0, ErrRpcUnknown
	}
	if resp.GetErr() != nil {
		return 0, resp.GetErr()
	}
	return int32(resp.GetState()), nil
}

func (s *GfSpClient) GetGroupList(
	ctx context.Context,
	name string,
	prefix string,
	sourceType string,
	limit int64,
	offset int64,
	includeRemoved bool,
	opts ...grpc.DialOption) ([]*types.Group, int64, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, 0, ErrRpcUnknown
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
		return nil, 0, ErrRpcUnknown
	}
	return resp.Groups, resp.Count, nil
}

func (s *GfSpClient) ListBucketsByBucketID(ctx context.Context, bucketIDs []uint64, includeRemoved bool, opts ...grpc.DialOption) (map[uint64]*types.Bucket, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListBucketsByBucketIDRequest{
		BucketIds:      bucketIDs,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListBucketsByBucketID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list buckets by bucket ids", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Buckets, nil
}

func (s *GfSpClient) ListObjectsByObjectID(ctx context.Context, objectIDs []uint64, includeRemoved bool, opts ...grpc.DialOption) (map[uint64]*types.Object, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListObjectsByObjectIDRequest{
		ObjectIds:      objectIDs,
		IncludeRemoved: includeRemoved,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListObjectsByObjectID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by object ids", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) ListVirtualGroupFamiliesSpID(ctx context.Context, spID uint32, opts ...grpc.DialOption) ([]*types.GlobalVirtualGroupFamily, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListVirtualGroupFamiliesBySpIDRequest{SpId: spID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListVirtualGroupFamiliesBySpID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list virtual group families by sp id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.GlobalVirtualGroupFamilies, nil
}

func (s *GfSpClient) GetGlobalVirtualGroupByGvgID(ctx context.Context, gvgID uint32, opts ...grpc.DialOption) (*types.GlobalVirtualGroup, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetGlobalVirtualGroupByGvgIDRequest{GvgId: gvgID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetGlobalVirtualGroupByGvgID(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get global virtual group by gvg id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.GlobalVirtualGroup, nil
}

func (s *GfSpClient) GetVirtualGroupFamilyBindingOnBucket(ctx context.Context, bucketID uint64, opts ...grpc.DialOption) (*types.GlobalVirtualGroupFamily, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetVirtualGroupFamilyBindingOnBucketRequest{BucketId: bucketID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetVirtualGroupFamilyBindingOnBucket(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get virtual group family binding on bucket", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.GlobalVirtualGroupFamily, nil
}

func (s *GfSpClient) ListBucketsBindingOnPrimarySP(ctx context.Context, spID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.Bucket, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListBucketsBindingOnPrimarySPRequest{
		SpId:       spID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListBucketsBindingOnPrimarySP(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list buckets by primary sp id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Buckets, nil
}

func (s *GfSpClient) ListBucketsBindingOnSecondarySP(ctx context.Context, spID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.Bucket, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListBucketsBindingOnSecondarySPRequest{
		SpId:       spID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListBucketsBindingOnSecondarySP(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list buckets by secondary sp id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Buckets, nil
}

func (s *GfSpClient) ListPrimaryObjects(ctx context.Context, spID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.Object, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListPrimaryObjectsRequest{
		SpId:       spID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListPrimaryObjects(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by primary sp id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) ListSecondaryObjects(ctx context.Context, spID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.Object, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListSecondaryObjectsRequest{
		SpId:       spID,
		StartAfter: startAfter,
		Limit:      limit,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListSecondaryObjects(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list objects by secondary sp id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) ListObjectsInGVG(ctx context.Context, gvgID uint32, startAfter uint64, limit uint32, opts ...grpc.DialOption) ([]*types.Object, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
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
		return nil, ErrRpcUnknown
	}
	return resp.Objects, nil
}

func (s *GfSpClient) GfSpGetVirtualGroupFamily(ctx context.Context, vgfID uint32, opts ...grpc.DialOption) (*types.GlobalVirtualGroupFamily, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetVirtualGroupFamilyRequest{VgfId: vgfID}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetVirtualGroupFamily(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get global virtual group family by vgf id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Vgf, nil
}

func (s *GfSpClient) GfSpGetGlobalVirtualGroup(ctx context.Context, bucketID uint64, lvgID uint32, opts ...grpc.DialOption) (*types.GlobalVirtualGroup, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpGetGlobalVirtualGroupRequest{
		BucketId: bucketID,
		LvgId:    lvgID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpGetGlobalVirtualGroup(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to get global virtual group by lvg id and bucket id", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Gvg, nil
}

func (s *GfSpClient) GfSpListMigrateBucketsEvent(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) ([]*types.EventMigrationBucket, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListMigrateBucketEventsRequest{
		BlockId: blockID,
		SpId:    spID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListMigrateBucketEvents(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list migrate bucket events", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Events, nil
}

func (s *GfSpClient) GfSpListSwapOutEvents(ctx context.Context, blockID uint64, spID uint32, opts ...grpc.DialOption) ([]*types.EventSwapOut, error) {
	conn, connErr := s.Connection(ctx, s.metadataEndpoint, opts...)
	if connErr != nil {
		log.CtxErrorw(ctx, "client failed to connect metadata", "error", connErr)
		return nil, ErrRpcUnknown
	}
	defer conn.Close()
	req := &types.GfSpListSwapOutEventsRequest{
		BlockId: blockID,
		SpId:    spID,
	}
	resp, err := types.NewGfSpMetadataServiceClient(conn).GfSpListSwapOutEvents(ctx, req)
	if err != nil {
		log.CtxErrorw(ctx, "client failed to list migrate swap out events", "error", err)
		return nil, ErrRpcUnknown
	}
	return resp.Events, nil
}
