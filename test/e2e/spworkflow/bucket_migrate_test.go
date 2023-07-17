package spworkflow

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	types "github.com/wealdtech/go-eth2-wallet-types/v2"

	sdkTypes "github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/bnb-chain/greenfield-storage-provider/test/e2e/spworkflow/basesuite"
	storageTestUtil "github.com/bnb-chain/greenfield/testutil/storage"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	types3 "github.com/bnb-chain/greenfield/x/sp/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type BucketMigrateTestSuite struct {
	basesuite.BaseSuite
	PrimarySP spTypes.StorageProvider

	SPList []spTypes.StorageProvider

	// destSP config
	OperatorAcc *types.Account
	FundingAcc  *types.Account
	SealAcc     *types.Account
	ApprovalAcc *types.Account
	GcAcc       *types.Account
	BlsAcc      *types.Account
}

func (s *BucketMigrateTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	spList, err := s.Client.ListStorageProviders(s.ClientContext, false)
	s.Require().NoError(err)
	for _, sp := range spList {
		if sp.Endpoint != "https://sp0.greenfield.io" {
			s.PrimarySP = sp
		}
	}
	s.SPList = spList
}

func TestBucketMigrateTestSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(BucketMigrateTestSuite))
}

func (s *BucketMigrateTestSuite) CreateObjects(bucketName string, count int) ([]*sdkTypes.ObjectDetail, []bytes.Buffer, error) {
	var (
		objectNames   []string
		contentBuffer []bytes.Buffer
		objectDetails []*sdkTypes.ObjectDetail
	)

	// create object
	for i := 0; i < count; i++ {
		var buffer bytes.Buffer
		line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,123456789012`
		// Create 1MiB content where each line contains 1024 characters.
		for i := 0; i < 1024*3; i++ {
			buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
		}
		objectName := storageTestUtil.GenRandomObjectName()
		s.T().Logf("---> CreateObject and HeadObject, bucketname:%s, objectname:%s <---", bucketName, objectName)
		objectTx, err := s.Client.CreateObject(s.ClientContext, bucketName, objectName, bytes.NewReader(buffer.Bytes()), sdkTypes.CreateObjectOptions{})
		s.Require().NoError(err)
		_, err = s.Client.WaitForTx(s.ClientContext, objectTx)
		s.Require().NoError(err)

		objectNames = append(objectNames, objectName)
		contentBuffer = append(contentBuffer, buffer)
	}

	// head object
	time.Sleep(5 * time.Second)
	for _, objectName := range objectNames {
		objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
		s.Require().NoError(err)
		s.Require().Equal(objectDetail.ObjectInfo.ObjectName, objectName)
		s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_CREATED")
	}

	s.T().Log("---> PutObject and GetObject <---")
	for idx, objectName := range objectNames {
		buffer := contentBuffer[idx]
		err := s.Client.PutObject(s.ClientContext, bucketName, objectName, int64(buffer.Len()),
			bytes.NewReader(buffer.Bytes()), sdkTypes.PutObjectOptions{})
		s.Require().NoError(err)
	}

	time.Sleep(20 * time.Second)
	// seal object
	for idx, objectName := range objectNames {
		objectDetail, err := s.Client.HeadObject(s.ClientContext, bucketName, objectName)
		s.Require().NoError(err)
		s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")

		ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectName, sdkTypes.GetObjectOptions{})
		s.Require().NoError(err)
		if err == nil {
			s.Require().Equal(info.ObjectName, objectName)
			objectBytes, err := io.ReadAll(ior)
			s.Require().NoError(err)
			s.Require().Equal(objectBytes, contentBuffer[idx].Bytes())
		}
		objectDetails = append(objectDetails, objectDetail)
	}

	return objectDetails, contentBuffer, nil
}

// test only one object's case
func (s *BucketMigrateTestSuite) Test_Bucket_Migrate_Simple_Case() {
	bucketName := storageTestUtil.GenRandomBucketName()

	// 1) create bucket and object in srcSP
	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, sdkTypes.CreateBucketOptions{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
	}

	// test only one object's case
	objectDetails, contentBuffer, err := s.CreateObjects(bucketName, 1)
	s.Require().NoError(err)

	objectDetail := objectDetails[0]
	buffer := contentBuffer[0]

	// selete a storage provider to miragte
	sps, err := s.Client.ListStorageProviders(s.ClientContext, true)
	s.Require().NoError(err)

	spIDs := make(map[uint32]bool)
	spIDs[objectDetail.GlobalVirtualGroup.PrimarySpId] = true
	for _, id := range objectDetail.GlobalVirtualGroup.SecondarySpIds {
		spIDs[id] = true
	}
	s.Require().Equal(len(spIDs), 7)

	var destSP *types3.StorageProvider
	for _, sp := range sps {
		_, exist := spIDs[sp.Id]
		if !exist {
			destSP = &sp
			break
		}
	}
	s.Require().NotNil(destSP)

	s.T().Logf(":Migrate Bucket DstPrimarySPID %d", destSP.GetId())

	// normal no conflict send migrate bucket transaction
	txhash, err := s.Client.MigrateBucket(s.ClientContext, bucketName, sdkTypes.MigrateBucketOptions{TxOpts: nil, DstPrimarySPID: destSP.GetId(), IsAsyncMode: false})
	s.Require().NoError(err)

	s.T().Logf("MigrateBucket : %s", txhash)

	for {
		bucketInfo, err = s.Client.HeadBucket(s.ClientContext, bucketName)
		s.T().Logf("HeadBucket: %s", bucketInfo)
		s.Require().NoError(err)
		if bucketInfo.BucketStatus != storageTypes.BUCKET_STATUS_MIGRATING {
			break
		}
		time.Sleep(3 * time.Second)
	}

	family, err := s.Client.QueryVirtualGroupFamily(s.ClientContext, bucketInfo.GlobalVirtualGroupFamilyId)
	s.Require().NoError(err)
	s.Require().Equal(family.PrimarySpId, destSP.GetId())
	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectDetail.ObjectInfo.ObjectName, sdkTypes.GetObjectOptions{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectDetail.ObjectInfo.ObjectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}
}

// test only conflict sp's case
func (s *BucketMigrateTestSuite) Test_Bucket_Migrate_Simple_Conflict_Case() {
	bucketName := storageTestUtil.GenRandomBucketName()

	// 1) create bucket and object in srcSP
	bucketTx, err := s.Client.CreateBucket(s.ClientContext, bucketName, s.PrimarySP.OperatorAddress, sdkTypes.CreateBucketOptions{})
	s.Require().NoError(err)

	_, err = s.Client.WaitForTx(s.ClientContext, bucketTx)
	s.Require().NoError(err)

	bucketInfo, err := s.Client.HeadBucket(s.ClientContext, bucketName)
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(bucketInfo.Visibility, storageTypes.VISIBILITY_TYPE_PRIVATE)
	}

	// test only one object's case
	objectDetails, contentBuffer, err := s.CreateObjects(bucketName, 1)
	s.Require().NoError(err)

	objectDetail := objectDetails[0]
	buffer := contentBuffer[0]

	// selete a storage provider to miragte
	sps, err := s.Client.ListStorageProviders(s.ClientContext, true)
	s.Require().NoError(err)

	spIDs := make(map[uint32]bool)
	spIDs[objectDetail.GlobalVirtualGroup.PrimarySpId] = true
	for _, id := range objectDetail.GlobalVirtualGroup.SecondarySpIds {
		spIDs[id] = true
	}
	s.Require().Equal(len(spIDs), 7)

	var destSP *types3.StorageProvider
	for _, sp := range sps {
		_, exist := spIDs[sp.Id]
		if !exist {
			destSP = &sp
			break
		}
	}
	s.Require().NotNil(destSP)

	// migrate bucket with conflict
	conflictSPID := objectDetail.GlobalVirtualGroup.SecondarySpIds[0]
	s.T().Logf(":Migrate Bucket DstPrimarySPID %d", conflictSPID)

	txhash, err := s.Client.MigrateBucket(s.ClientContext, bucketName, sdkTypes.MigrateBucketOptions{TxOpts: nil, DstPrimarySPID: conflictSPID, IsAsyncMode: false})
	s.Require().NoError(err)

	s.T().Logf("MigrateBucket : %s", txhash)

	for {
		bucketInfo, err = s.Client.HeadBucket(s.ClientContext, bucketName)
		s.T().Logf("HeadBucket: %s", bucketInfo)
		s.Require().NoError(err)
		if bucketInfo.BucketStatus != storageTypes.BUCKET_STATUS_MIGRATING {
			break
		}
		time.Sleep(3 * time.Second)
	}

	family, err := s.Client.QueryVirtualGroupFamily(s.ClientContext, bucketInfo.GlobalVirtualGroupFamilyId)
	s.Require().NoError(err)
	s.Require().Equal(family.PrimarySpId, conflictSPID)
	ior, info, err := s.Client.GetObject(s.ClientContext, bucketName, objectDetail.ObjectInfo.ObjectName, sdkTypes.GetObjectOptions{})
	s.Require().NoError(err)
	if err == nil {
		s.Require().Equal(info.ObjectName, objectDetail.ObjectInfo.ObjectName)
		objectBytes, err := io.ReadAll(ior)
		s.Require().NoError(err)
		s.Require().Equal(objectBytes, buffer.Bytes())
	}
}
