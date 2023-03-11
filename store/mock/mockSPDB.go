package mock

import (
	reflect "reflect"
	"time"

	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/golang/mock/gomock"
)

// MockSPDB is a mock of SPDB interface.
type MockSPDB struct {
	ctrl     *gomock.Controller
	recorder *MockSPDBMockRecorder
}

func (m *MockSPDB) CreateUploadJob(objectInfo *storagetypes.ObjectInfo) (*servicetypes.JobContext, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) UpdateJobState(objectID uint64, state servicetypes.JobState) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetJobByID(jobID uint64) (*servicetypes.JobContext, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetJobByObjectID(objectID uint64) (*servicetypes.JobContext, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetObjectInfo(objectID uint64) (*storagetypes.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) SetObjectInfo(objectID uint64, objectInfo *storagetypes.ObjectInfo) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetObjectIntegrity(objectID uint64) (*sqldb.IntegrityMeta, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) SetObjectIntegrity(integrity *sqldb.IntegrityMeta) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) CheckQuotaAndAddReadRecord(record *sqldb.ReadRecord, quota *sqldb.BucketQuota) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetBucketTraffic(bucketID uint64, yearMonth string) (*sqldb.BucketTraffic, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetReadRecord(timeRange *sqldb.TrafficTimeRange) ([]*sqldb.ReadRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetBucketReadRecord(bucketID uint64, timeRange *sqldb.TrafficTimeRange) ([]*sqldb.ReadRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetObjectReadRecord(objectID uint64, timeRange *sqldb.TrafficTimeRange) ([]*sqldb.ReadRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetUserReadRecord(userAddress string, timeRange *sqldb.TrafficTimeRange) ([]*sqldb.ReadRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) UpdateAllSp(spList []*sptypes.StorageProvider) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) FetchAllSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) FetchAllSpWithoutOwnSp(status ...sptypes.Status) ([]*sptypes.StorageProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetSpByAddress(address string, addressType sqldb.SpAddressType) (*sptypes.StorageProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetSpByEndpoint(endpoint string) (*sptypes.StorageProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetOwnSpInfo() (*sptypes.StorageProvider, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) SetOwnSpInfo(sp *sptypes.StorageProvider) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) GetStorageParams() (*storagetypes.Params, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) SetStorageParams(params *storagetypes.Params) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockSPDB) InsertAuthKey(newRecord *sqldb.OffChainAuthKeyTable) error {
	//TODO implement me
	panic("implement me")
}

// MockSPDBMockRecorder is the mock recorder for MockSPDB.
type MockSPDBMockRecorder struct {
	mock *MockSPDB
}

// NewMockSPDB creates a new mock instance.
func NewMockSPDB(ctrl *gomock.Controller) *MockSPDB {
	mock := &MockSPDB{ctrl: ctrl}
	mock.recorder = &MockSPDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSPDB) EXPECT() *MockSPDBMockRecorder {
	return m.recorder
}

// GetAuthKey mocks base method.
func (m *MockSPDB) GetAuthKey(userAddress string, domain string) (*sqldb.OffChainAuthKeyTable, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuthKey", userAddress, domain)
	ret0, _ := ret[0].(*sqldb.OffChainAuthKeyTable)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAuthKey indicates an expected call of GetAuthKey.
func (mr *MockSPDBMockRecorder) GetAuthKey(userAddress, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuthKey", reflect.TypeOf((*MockSPDB)(nil).GetAuthKey), userAddress, domain)
}

// UpdateAuthKey mocks base method.
func (m *MockSPDB) UpdateAuthKey(userAddress string, domain string, oldNonce int32, newNonce int32, newPublicKey string, newExpiryDate time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAuthKey", userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAuthKey indicates an expected call of UpdateAuthKey.
func (mr *MockSPDBMockRecorder) UpdateAuthKey(userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAuthKey", reflect.TypeOf((*MockSPDB)(nil).UpdateAuthKey), userAddress, domain, oldNonce, newNonce, newPublicKey, newExpiryDate)
}
