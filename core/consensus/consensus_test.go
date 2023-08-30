package consensus

import (
	"context"
	"testing"
)

func TestNullConsensus(t *testing.T) {
	nc := &NullConsensus{}
	_, _ = nc.CurrentHeight(context.TODO())
	_, _ = nc.HasAccount(context.TODO(), "")
	_, _ = nc.ListSPs(context.TODO())
	_, _ = nc.QuerySP(context.TODO(), "")
	_, _ = nc.QuerySPFreeQuota(context.TODO(), "")
	_, _ = nc.QuerySPPrice(context.TODO(), "")
	_, _ = nc.QuerySPByID(context.TODO(), 0)
	_, _ = nc.ListBondedValidators(context.TODO())
	_, _ = nc.ListVirtualGroupFamilies(context.TODO(), 0)
	_, _ = nc.ListGlobalVirtualGroupsByFamilyID(context.TODO(), 0)
	_, _ = nc.AvailableGlobalVirtualGroupFamilies(context.TODO(), []uint32{})
	_, _ = nc.QueryVirtualGroupFamily(context.TODO(), 0)
	_, _ = nc.QueryGlobalVirtualGroup(context.TODO(), 0)
	_, _ = nc.QueryVirtualGroupParams(context.TODO())
	_, _ = nc.QueryStorageParams(context.TODO())
	_, _ = nc.QueryStorageParamsByTimestamp(context.TODO(), 0)
	_, _ = nc.QueryBucketInfo(context.TODO(), "")
	_, _ = nc.QueryBucketInfoById(context.TODO(), 0)
	_, _ = nc.QueryObjectInfo(context.TODO(), "", "")
	_, _ = nc.QueryObjectInfoByID(context.TODO(), "")
	_, _, _ = nc.QueryBucketInfoAndObjectInfo(context.TODO(), "", "")
	_, _ = nc.QueryPaymentStreamRecord(context.TODO(), "")
	_, _ = nc.VerifyGetObjectPermission(context.TODO(), "", "", "")
	_, _ = nc.VerifyPutObjectPermission(context.TODO(), "", "", "")
	_, _ = nc.ListenObjectSeal(context.TODO(), 0, 0)
	_, _ = nc.ListenRejectUnSealObject(context.TODO(), 0, 0)
	_, _ = nc.ConfirmTransaction(context.TODO(), "")
	_ = nc.WaitForNextBlock(context.TODO())
	_ = nc.Close()
}
