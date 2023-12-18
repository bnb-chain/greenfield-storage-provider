package test

import (
	"errors"
	"fmt"
	golog "log"
	"math/big"
	"testing"
	"time"

	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/models"
	"github.com/shopspring/decimal"
	"github.com/spaolacci/murmur3"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var verifyFuncs = []func(t *testing.T, db *gorm.DB) error{
	verify1, verify2, verify3, verify4, verify5, verify6, verify7, verify8, verify9, verify10,
	verify11, verify12, verify13, verify14, verify15, verify16, verify17, verify18, verify19, verify20,
	verify21, verify22, verify23, verify24, verify25, verify26, verify27, verify28, verify29, verify30,
	verify31, verify32, verify33, verify34, verify35, verify36, verify37, verify38, verify39, verify40,
	verify41, verify42, verify43, verify44, verify45, verify46, verify47, verify48, verify49, verify50,
	verify51, verify52, verify53, verify54, verify55, verify56, verify57, verify58, verify59, verify60,
	verify61, verify62, verify63,
}

func Verify(t *testing.T) error {
	dsn := "root:root@tcp(localhost:3306)/block_syncer?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	tick := time.NewTicker(time.Millisecond * 500)
	defer tick.Stop()
	for i := 1; i <= LatestHeight; i++ {
		tryCount := 0
		for tryCount < 10 {
			<-tick.C
			var epoch bsdb.Epoch
			db.Table("epoch").First(&epoch)
			tryCount++
			if epoch.BlockHeight == int64(i) {
				break
			}
		}

		// verify data
		f := verifyFuncs[i-1]
		if err := f(t, db); err != nil {
			return err
		}

		golog.Printf("%d case pass", i)

		// height increase
		StatusRes = fmt.Sprintf("{\"sync_info\":{\"latest_block_height\":\"%d\"}}", i+1)
	}

	// verify height
	var epoch bsdb.Epoch
	db.Table("epoch").First(&epoch)
	if epoch.BlockHeight != int64(LatestHeight) {
		return errors.New("height error")
	}

	return nil
}

func verify1(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "full-node-v1-acc0000000072-buc0000000000").Find(&bucket).Error; err != nil {
		return err
	}
	golog.Println(bucket)
	if bucket.Status != "BUCKET_STATUS_CREATED" {
		return errors.New("bucket status error")
	}
	return nil
}

func verify2(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "full-node-v1-acc0000000072-buc0000000000").Find(&bucket).Error; err != nil {
		return err
	}
	if bucket.Visibility != "VISIBILITY_TYPE_PUBLIC_READ" {
		return errors.New("bucket Visibility is error")
	}
	return nil
}
func verify3(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "full-node-v1-acc0000000072-buc0000000000").Find(&bucket).Error; err != nil {
		return err
	}
	if !bucket.Removed {
		return errors.New("bucket is not remove")
	}
	return nil
}
func verify4(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "cxz").Find(&bucket).Error; err != nil {
		return err
	}
	if bucket.DeleteAt != int64(4) {
		return fmt.Errorf("bucket is not discontinue delete at:%v", bucket.DeleteAt)
	}
	return nil
}
func verify5(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "cxz").Find(&bucket).Error; err != nil {
		return err
	}
	if bucket.GlobalVirtualGroupFamilyId != 9 {
		return errors.New("bucket GlobalVirtualGroupFamilyId error")
	}
	return nil
}
func verify6(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "cxz").Find(&bucket).Error; err != nil {
		return err
	}
	if bucket.GlobalVirtualGroupFamilyId != 10 {
		return errors.New("bucket GlobalVirtualGroupFamilyId error")
	}
	return nil
}
func verify7(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")).Find(&g).Error; err != nil {
		return err
	}
	return nil
}
func verify8(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ? and account_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"), common.HexToAddress("0x92b7976702C064C7e2e791854497Ec73C853CEB5")).Find(&g).Error; err != nil {
		return err
	}
	if g.ExpirationTime != 253402300799 {
		return errors.New("member ExpirationTime error")
	}
	return nil
}
func verify9(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ? and account_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"), common.HexToAddress("0x4d57d300AfaF9f407e26552965ce355786206cF4")).Find(&g).Error; err != nil {
		return err
	}
	if g.ExpirationTime != 1704067199 {
		return fmt.Errorf("member ExpirationTime error ExpirationTime:%d account id:%v", g.ExpirationTime, g.AccountID)
	}
	return nil
}
func verify10(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ? and account_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"), common.HexToAddress("0x0000000000000000000000000000000000000000")).Find(&g).Error; err != nil {
		return err
	}
	if !g.Removed {
		return errors.New("group is not removed")
	}
	return nil
}
func verify11(t *testing.T, db *gorm.DB) error {
	table := GetObjectsTableName("cxz")
	var o models.Object
	if err := db.Table(table).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&o).Error; err != nil {
		return err
	}
	return nil
}
func verify12(t *testing.T, db *gorm.DB) error {
	table := GetObjectsTableName("cxz")
	var o models.Object
	if err := db.Table(table).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&o).Error; err != nil {
		return err
	}
	if o.Status != "OBJECT_STATUS_SEALED" {
		return errors.New("object status error")
	}
	var b models.Bucket
	if err := db.Table("buckets").Where("bucket_name =?", "cxz").Find(&b).Error; err != nil {
		return err
	}
	if !b.StorageSize.Equal(decimal.NewFromInt(int64(o.PayloadSize))) {
		return fmt.Errorf("StorageSize: %s error", b.StorageSize.String())
	}
	if !b.ChargeSize.Equal(decimal.NewFromInt(128000)) {
		return fmt.Errorf("ChargeSize: %s error", b.ChargeSize.String())
	}
	return nil
}
func verify13(t *testing.T, db *gorm.DB) error {
	table := GetObjectsTableName("cxz")
	var o models.Object
	if err := db.Table(table).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&o).Error; err != nil {
		return err
	}
	if !o.Removed {
		return errors.New("object is not removed")
	}
	return nil
}
func verify14(t *testing.T, db *gorm.DB) error {
	var sr models.StreamRecord
	if err := db.Table((&models.StreamRecord{}).TableName()).Where("account = ?", common.HexToAddress("0x51D2c7dFa2a58C73853c2657Dc505e7430879B1b")).Find(&sr).Error; err != nil {
		return err
	}
	return nil
}
func verify15(t *testing.T, db *gorm.DB) error {
	var p models.PaymentAccount
	if err := db.Table((&models.PaymentAccount{}).TableName()).Where("addr = ?", common.HexToAddress("0x68bd245Df652435321989b999F9F70Cd31281b66")).Find(&p).Error; err != nil {
		return err
	}
	return nil
}
func verify16(t *testing.T, db *gorm.DB) error {
	var lvg models.LocalVirtualGroup
	if err := db.Table((&models.LocalVirtualGroup{}).TableName()).Where("local_virtual_group_id = ?", 8).Find(&lvg).Error; err != nil {
		return err
	}
	return nil
}
func verify17(t *testing.T, db *gorm.DB) error {
	var lvg models.LocalVirtualGroup
	if err := db.Table((&models.LocalVirtualGroup{}).TableName()).Where("local_virtual_group_id = ?", 8).Find(&lvg).Error; err != nil {
		return err
	}
	if lvg.StoredSize != 0 {
		return errors.New("StoredSize error")
	}
	return nil
}
func verify18(t *testing.T, db *gorm.DB) error {
	var gvg bsdb.GlobalVirtualGroup
	if err := db.Table((&bsdb.GlobalVirtualGroup{}).TableName()).Where("global_virtual_group_id = ?", 1).Find(&gvg).Error; err != nil {
		return err
	}
	return nil
}
func verify19(t *testing.T, db *gorm.DB) error {
	var gvg bsdb.GlobalVirtualGroup
	if err := db.Table((&bsdb.GlobalVirtualGroup{}).TableName()).Where("global_virtual_group_id = ?", 1).Find(&gvg).Error; err != nil {
		return err
	}
	if gvg.StoredSize != 48372064 || gvg.TotalDeposit.Raw().String() != "140737488355328000" {
		return fmt.Errorf("gvg update error StoredSize: %v, TotalDeposit:%v", gvg.StoredSize, gvg.TotalDeposit.Raw().String())
	}
	return nil
}
func verify20(t *testing.T, db *gorm.DB) error {
	var gvgf bsdb.GlobalVirtualGroupFamily
	if err := db.Table((&bsdb.GlobalVirtualGroupFamily{}).TableName()).Where("global_virtual_group_family_id = ?", 1).Find(&gvgf).Error; err != nil {
		return err
	}
	return nil
}
func verify21(t *testing.T, db *gorm.DB) error {
	var gvgf bsdb.GlobalVirtualGroupFamily
	if err := db.Table((&bsdb.GlobalVirtualGroupFamily{}).TableName()).Where("global_virtual_group_family_id = ?", 1).Find(&gvgf).Error; err != nil {
		return err
	}
	if gvgf.PrimarySpId != 3 {
		return errors.New("PrimarySpId error")
	}
	return nil
}
func verify22(t *testing.T, db *gorm.DB) error {
	var gvgf bsdb.GlobalVirtualGroupFamily
	if err := db.Table((&bsdb.GlobalVirtualGroupFamily{}).TableName()).Where("global_virtual_group_family_id = ?", 1).Find(&gvgf).Error; err != nil {
		return err
	}
	if !gvgf.Removed {
		return errors.New("gvgf is not removed")
	}
	return nil
}
func verify23(t *testing.T, db *gorm.DB) error {
	var sp models.StorageProvider
	if err := db.Table((&models.StorageProvider{}).TableName()).Where("sp_id = ?", 14).Find(&sp).Error; err != nil {
		return err
	}
	return nil
}
func verify24(t *testing.T, db *gorm.DB) error {
	var sp models.StorageProvider
	if err := db.Table((&models.StorageProvider{}).TableName()).Where("sp_id = ?", 14).Find(&sp).Error; err != nil {
		return err
	}
	if sp.Endpoint != "http://spxrmfl.greenfield.io" || sp.BlsKey != "b689357b256f8aabaf02fceb56a9a61c59b2d9b3cc78d4413fefd1e3bd902c90dcc7b346deb672d164c9ff832a8ee1d9" {
		return fmt.Errorf("sp update failed endpoint: %s, blskey:%s", sp.Endpoint, sp.BlsKey)
	}
	return nil
}
func verify25(t *testing.T, db *gorm.DB) error {
	var sp models.StorageProvider
	if err := db.Table((&models.StorageProvider{}).TableName()).Where("sp_id = ?", 14).Find(&sp).Error; err != nil {
		return err
	}
	if !sp.Removed {
		return errors.New("sp is not removed")
	}
	return nil
}

func verify26(t *testing.T, db *gorm.DB) error {
	var permission models.Permission
	if err := db.Table((&models.Permission{}).TableName()).Where("policy_id = ?", common.BigToHash(big.NewInt(2))).Find(&permission).Error; err != nil {
		return err
	}
	return nil
}

func verify27(t *testing.T, db *gorm.DB) error {
	var permission models.Permission
	if err := db.Table((&models.Permission{}).TableName()).Where("policy_id = ?", common.BigToHash(big.NewInt(2))).Find(&permission).Error; err != nil {
		return err
	}
	if !permission.Removed {
		return errors.New("permission is not removed")
	}
	return nil
}

func verify28(t *testing.T, db *gorm.DB) error {
	var object models.Object
	if err := db.Table(GetObjectsTableName("cxz")).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&object).Error; err != nil {
		return err
	}
	if !object.Removed {
		return errors.New("object is not removed")
	}
	return nil
}
func verify29(t *testing.T, db *gorm.DB) error {
	var object models.Object
	if err := db.Table(GetObjectsTableName("cxz")).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&object).Error; err != nil {
		return err
	}
	if object.Status != storagetypes.OBJECT_STATUS_DISCONTINUED.String() {
		return errors.New("object status error")
	}
	return nil
}
func verify30(t *testing.T, db *gorm.DB) error {
	return nil
}
func verify31(t *testing.T, db *gorm.DB) error {
	var object models.Object
	if err := db.Table(GetObjectsTableName("cxz")).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&object).Error; err != nil {
		return err
	}
	if !object.Removed || object.Operator.String() != "0x6fD578b6fd9635cB7a7dDF53B13DbB0c873aEFCD" {
		return errors.New("object is not removed")
	}
	return nil
}
func verify32(t *testing.T, db *gorm.DB) error {
	var object models.Object
	if err := db.Table(GetObjectsTableName("cxz")).Where("object_id = ?", common.BigToHash(big.NewInt(1201))).Find(&object).Error; err != nil {
		return err
	}
	if object.Visibility != "VISIBILITY_TYPE_INHERIT" {
		return errors.New("object Visibility error")
	}
	return nil
}
func verify33(t *testing.T, db *gorm.DB) error {
	var object models.Object
	if err := db.Table(GetObjectsTableName("cxz")).Where("object_id = ?", common.BigToHash(big.NewInt(1204))).Find(&object).Error; err != nil {
		return err
	}
	return nil
}
func verify34(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventCompleteMigrationBucket
	if err := db.Table(bsdb.EventCompleteMigrationTableName).Where("bucket_id = ?", common.BigToHash(big.NewInt(4))).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify35(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventMigrationBucket
	if err := db.Table(bsdb.EventMigrationTableName).Where("bucket_id = ?", common.BigToHash(big.NewInt(4))).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify36(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventCancelMigrationBucket
	if err := db.Table(bsdb.EventCancelMigrationTableName).Where("bucket_id = ?", common.BigToHash(big.NewInt(4))).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify37(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventStorageProviderExit
	if err := db.Table(bsdb.EventStorageProviderExitTableName).Where("storage_provider_id = ?", 9).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify38(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventSwapOut
	if err := db.Table(bsdb.EventSwapOutTableName).Where("storage_provider_id = ?", 9).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify39(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventCompleteSwapOut
	if err := db.Table(bsdb.EventCompleteSwapOutTableName).Where("storage_provider_id = ?", 9).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify40(t *testing.T, db *gorm.DB) error {
	var cmb bsdb.EventCompleteStorageProviderExit
	if err := db.Table(bsdb.EventCompleteStorageProviderExitTableName).Where("storage_provider_id = ?", 10).Find(&cmb).Error; err != nil {
		return err
	}
	return nil
}
func verify41(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table((&models.Group{}).TableName()).Where("account_id = ? and group_id = ?", common.HexToAddress("0x4d57d300AfaF9f407e26552965ce355786206cF4"), common.HexToHash("2")).Find(&g).Error; err != nil {
		return err
	}
	if g.ExpirationTime != 33260975999 {
		return errors.New("ExpirationTime err")
	}
	return nil
}
func verify42(t *testing.T, db *gorm.DB) error {
	var sp models.StorageProvider
	if err := db.Table((&models.StorageProvider{}).TableName()).Where("sp_id = ?", 14).Find(&sp).Error; err != nil {
		return err
	}
	if sp.ReadPrice.Raw().Cmp(big.NewInt(1000.0)) == 0 || sp.StorePrice.Raw().Cmp(big.NewInt(100000)) == 0 {
		return errors.New("price error")
	}
	return nil
}

func verify43(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000016")).Find(&g).Error; err != nil {
		return err
	}
	return nil
}

func verify44(t *testing.T, db *gorm.DB) error {
	return nil
}

func verify45(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ? and account_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000016"), common.HexToAddress("0x5870Af236E63beaEbbFa364f78FC7c8e70F0811f")).Find(&g).Error; err != nil {
		return err
	}
	if g.ExpirationTime != 1356998399 {
		return fmt.Errorf("member expiration time is error :%d", g.ExpirationTime)
	}
	return nil
}

func verify46(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_id = ? and account_id = ?", common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000016"), common.HexToAddress("0x5870AF236E63BEAEBBFA364F78FC7C8E70F0811F")).Find(&g).Error; err != nil {
		return err
	}
	if !g.Removed {
		return errors.New("member is not removed")
	}
	return nil
}

func verify47(t *testing.T, db *gorm.DB) error {
	var p models.PaymentAccount
	if err := db.Table((&models.PaymentAccount{}).TableName()).Where("addr = ?", common.HexToAddress("0x68bd245Df652435321989b999F9F70Cd31281b66")).Find(&p).Error; err != nil {
		return err
	}
	if !p.Refundable {
		return fmt.Errorf("failed to update Refundable")
	}
	return nil
}

func verify48(t *testing.T, db *gorm.DB) error {
	var p models.PaymentAccount
	if err := db.Table((&models.PaymentAccount{}).TableName()).Where("addr = ?", common.HexToAddress("0x68bd245Df652435321989b999F9F70Cd31281b66")).Find(&p).Error; err != nil {
		return err
	}
	if p.Refundable {
		return fmt.Errorf("failed to update Refundable")
	}
	return nil
}

func verify49(t *testing.T, db *gorm.DB) error {
	var event bsdb.EventRejectMigrateBucket
	if err := db.Table((&bsdb.EventCompleteMigrationBucket{}).TableName()).Where("bucket_name = ?", "onkz").Find(&event).Error; err != nil {
		return errors.New("event not found")
	}
	return nil
}

func verify50(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "b4tag02").Find(&bucket).Error; err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, "{\"tags\": [{\"key\": \"key1\", \"value\": \"value1\"}, {\"key\": \"key2\", \"value\": \"value2\"}]}", bucket.Tags.String())

	return nil
}

func verify51(t *testing.T, db *gorm.DB) error {
	// skip this test. Because block 51 created a bucket without setting tags and we set the tag in block 52. So, we will verify the result in verify52 method
	return nil
}

func verify52(t *testing.T, db *gorm.DB) error {
	var bucket models.Bucket
	if err := db.Table(bsdb.BucketTableName).Where("bucket_name = ?", "b4tag04").Find(&bucket).Error; err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, "{\"tags\": [{\"key\": \"key1\", \"value\": \"value1\"}, {\"key\": \"key2\", \"value\": \"value2\"}, {\"key\": \"tag04\", \"value\": \"tag04_value\"}]}", bucket.Tags.String())

	return nil
}

func verify53(t *testing.T, db *gorm.DB) error {
	table := GetObjectsTableName("ot005test-bucket")
	var o models.Object
	if err := db.Table(table).Where("bucket_name=? and object_name = ?", "ot005test-bucket", "ot005obj002").Find(&o).Error; err != nil {
		assert.NoError(t, err)
	}

	assert.Equal(t, "{\"tags\": [{\"key\": \"key1\", \"value\": \"value1\"}, {\"key\": \"key2\", \"value\": \"value2\"}]}", o.Tags.String())

	return nil
}

func verify54(t *testing.T, db *gorm.DB) error {
	// skip this test. Because block 54 created an object  without setting tags and we set the tag in block 55. So, we will verify the result in verify55 method
	return nil
}

func verify55(t *testing.T, db *gorm.DB) error {
	table := GetObjectsTableName("ot005test-bucket")
	var o models.Object
	if err := db.Table(table).Where("bucket_name=? and object_name = ?", "ot005test-bucket", "ot005obj004").Find(&o).Error; err != nil {
		assert.NoError(t, err)
	}

	assert.Equal(t, "{\"tags\": [{\"key\": \"key1\", \"value\": \"value1\"}, {\"key\": \"key2\", \"value\": \"value2\"}, {\"key\": \"tag04\", \"value\": \"tag04_value\"}]}", o.Tags.String())

	return nil
}

func verify56(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_name = ? and account_id = ?", "g001-02", common.HexToAddress("0")).Find(&g).Error; err != nil {
		return err
	}
	assert.Equal(t, "{\"tags\": [{\"key\": \"key1\", \"value\": \"value1\"}, {\"key\": \"key2\", \"value\": \"value2\"}]}", g.Tags.String())

	return nil
}

func verify57(t *testing.T, db *gorm.DB) error {
	// skip this test. Because block 57 created an group  without setting tags and we set the tag in block 58. So, we will verify the result in verify58 method
	return nil
}

func verify58(t *testing.T, db *gorm.DB) error {
	var g models.Group
	if err := db.Table(bsdb.GroupTableName).Where("group_name = ? and account_id = ?", "g001-04", common.HexToAddress("0")).Find(&g).Error; err != nil {
		return err
	}

	assert.Equal(t, "{\"tags\": [{\"key\": \"tag04\", \"value\": \"tag04_value\"}]}", g.Tags.String())

	return nil
}

func verify59(t *testing.T, db *gorm.DB) error {
	var count int64
	if err := db.Table(GetPrefixesTableName("cxz")).Where("bucket_name = ? and is_folder = ? and full_name = ?", "cxz", true, "/coco/").Count(&count).Error; err != nil {
		return errors.New("event not found")
	}

	// if count == 2 which means there exist repeat folder in the root case
	if count >= 2 {
		return fmt.Errorf("failed to batch create slash prefix tree node")
	}
	return nil
}

func verify60(t *testing.T, db *gorm.DB) error {
	var count int64
	if err := db.Table(GetPrefixesTableName("cxz")).Where("bucket_name = ? and is_folder = ? and path_name = ?", "cxz", true, "/coco/").Count(&count).Error; err != nil {
		return errors.New("event not found")
	}

	// if count > 0 which means the /coco/ folder didn't delete successfully
	if count > 0 {
		return fmt.Errorf("failed to batch delete slash prefix tree node")
	}
	return nil
}

func verify61(t *testing.T, db *gorm.DB) error {
	var count int64
	if err := db.Table(GetPrefixesTableName("cxz")).Where("bucket_name = ? and is_folder = ? and path_name = ?", "cxz", true, "/sp/").Count(&count).Error; err != nil {
		return errors.New("event not found")
	}

	if count != 0 {
		return fmt.Errorf("delete and create same object in same block")
	}
	return nil
}

func verify62(t *testing.T, db *gorm.DB) error {
	var count int64
	if err := db.Table(GetPrefixesTableName("cxz")).Where("bucket_name = ? and full_name = ?", "cxz", "/coco/data/123.txt").Count(&count).Error; err != nil {
		return errors.New("event not found")
	}

	if count != 1 {
		return fmt.Errorf("delete and create same object in same block")
	}
	return nil
}

func verify63(t *testing.T, db *gorm.DB) error {
	var count int64
	if err := db.Table(GetPrefixesTableName("cxz")).Where("bucket_name = ? and full_name = ?", "cxz", "/coco/data/123.txt").Count(&count).Error; err != nil {
		return errors.New("event not found")
	}

	if count != 1 {
		return fmt.Errorf("delete and create same object in same block")
	}
	return nil
}

const ObjectsNumberOfShards = 64
const ObjectTableName = "objects"

func GetObjectsTableName(bucketName string) string {
	return GetObjectsTableNameByShardNumber(int(GetObjectsShardNumberByBucketName(bucketName)))
}

func GetObjectsShardNumberByBucketName(bucketName string) uint32 {
	return murmur3.Sum32([]byte(bucketName)) % ObjectsNumberOfShards
}

func GetObjectsTableNameByShardNumber(shard int) string {
	return fmt.Sprintf("%s_%02d", ObjectTableName, shard)
}

const PrefixesNumberOfShards = 64
const PrefixTreeTableName = "slash_prefix_tree_nodes"

func GetPrefixesTableName(bucketName string) string {
	return GetPrefixesTableNameByShardNumber(int(GetPrefixesShardNumberByBucketName(bucketName)))
}

func GetPrefixesShardNumberByBucketName(bucketName string) uint32 {
	return murmur3.Sum32([]byte(bucketName)) % PrefixesNumberOfShards
}

func GetPrefixesTableNameByShardNumber(shard int) string {
	return fmt.Sprintf("%s_%02d", PrefixTreeTableName, shard)
}
