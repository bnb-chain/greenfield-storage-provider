package main

import (
	"os"
	"time"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func main() {
	cfg := &config.SQLDBConfig{
		User:     "root",
		Passwd:   "bfs-test",
		Address:  "127.0.0.1:3306",
		Database: "sp_db",
	}
	store, err := sqldb.NewSQLStore(cfg)
	if err != nil {
		log.Errorw("new sql store failed", "error", err)
	}
	if err = jobFunc(store); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if err = objectFunc(store); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if err = integrityFunc(store); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func jobFunc(store *sqldb.SQLDB) error {
	log.Info("CreateUploadJob")
	jobCtx, err := store.CreateUploadJob(&storagetypes.ObjectInfo{
		Owner:                "yyyy",
		BucketName:           "bucket1",
		ObjectName:           "object1",
		Id:                   math.NewUint(123),
		PayloadSize:          25,
		IsPublic:             false,
		ContentType:          "application/json",
		CreateAt:             100,
		ObjectStatus:         storagetypes.OBJECT_STATUS_INIT,
		RedundancyType:       storagetypes.REDUNDANCY_EC_TYPE,
		SourceType:           storagetypes.SOURCE_TYPE_ORIGIN,
		Checksums:            [][]byte{[]byte("a")},
		SecondarySpAddresses: []string{"secondary_sp1"},
	})
	if err != nil {
		log.Errorw("CreateUploadJob failed", "error", err)
		return err
	}
	log.Infow("1", "jobCtx", jobCtx)

	log.Info("GetJobByID")
	jobCtx, err = store.GetJobByID(jobCtx.GetJobId())
	if err != nil {
		log.Errorw("GetJobByID failed", "error", err)
		return err
	}
	log.Infow("2", "jobCtx", jobCtx)

	log.Info("GetJobByObjectID")
	jobCtx, err = store.GetJobByObjectID(123)
	if err != nil {
		log.Errorw("GetJobByObjectID failed", "error", err)
		return err
	}
	log.Infow("3", "jobCtx", jobCtx)

	log.Info("UpdateJobState")
	err = store.UpdateJobState(123, servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DONE)
	if err != nil {
		log.Errorw("UpdateJobState failed", "error", err)
		return err
	}
	return nil
}

func objectFunc(store *sqldb.SQLDB) error {
	now := time.Now().Unix()
	log.Info("SetObjectInfo first")
	if err := store.SetObjectInfo(456, &storagetypes.ObjectInfo{
		Owner:                "zzz",
		BucketName:           "bucket2",
		ObjectName:           "object2",
		Id:                   math.NewUint(678),
		PayloadSize:          65,
		IsPublic:             true,
		ContentType:          "application/xml",
		CreateAt:             100,
		ObjectStatus:         storagetypes.OBJECT_STATUS_IN_SERVICE,
		RedundancyType:       storagetypes.REDUNDANCY_REPLICA_TYPE,
		SourceType:           storagetypes.SOURCE_TYPE_BSC_CROSS_CHAIN,
		Checksums:            [][]byte{[]byte("b")},
		SecondarySpAddresses: []string{"secondary_sp2"},
	}); err != nil {
		log.Errorw("SetObjectInfo failed", "error", err)
		return err
	}

	log.Info("GetObjectInfo")
	objectInfo, err := store.GetObjectInfo(456)
	if err != nil {
		log.Errorw("GetObjectInfo failed", "error", err)
		return err
	}
	log.Infow("4", "objectInfo", objectInfo)

	log.Info("SetObjectInfo second")
	if err := store.SetObjectInfo(456, &storagetypes.ObjectInfo{
		Owner:                "greenfield",
		BucketName:           "bucket3",
		ObjectName:           "object3",
		Id:                   math.NewUint(678),
		PayloadSize:          65,
		IsPublic:             true,
		ContentType:          "application/xml",
		CreateAt:             now,
		ObjectStatus:         storagetypes.OBJECT_STATUS_IN_SERVICE,
		RedundancyType:       storagetypes.REDUNDANCY_REPLICA_TYPE,
		SourceType:           storagetypes.SOURCE_TYPE_BSC_CROSS_CHAIN,
		Checksums:            [][]byte{[]byte("b")},
		SecondarySpAddresses: []string{"secondary_sp3"},
	}); err != nil {
		log.Errorw("SetObjectInfo failed", "error", err)
		return err
	}
	return nil
}

func integrityFunc(store *sqldb.SQLDB) error {
	log.Info("SetObjectIntegrity")
	if err := store.SetObjectIntegrity(&sqldb.IntegrityMeta{
		ObjectID:      561,
		Checksum:      [][]byte{[]byte("c")},
		IntegrityHash: []byte("c"),
		Signature:     []byte("d"),
	}); err != nil {
		log.Errorw("SetObjectIntegrity failed", "error", err)
		return err
	}

	log.Info("GetObjectIntegrity")
	integrity, err := store.GetObjectIntegrity(561)
	if err != nil {
		log.Errorw("GetObjectIntegrity failed", "error", err)
		return err
	}
	log.Infow("5", "integrity", integrity)
	return nil
}

func ownSPInfo(store *sqldb.SQLDB) error {
	log.Info("SetOwnSPInfo")
	if err := store.SetOwnSPInfo(&sptypes.StorageProvider{
		OperatorAddress: "0xD30717BFD945901946Ec74f1986cca3fcf59681B",
		FundingAddress:  "0xbE2B3F7318f288E898E15553C6B9c15B5D1F3A1e",
		SealAddress:     "0x29735289CDbda24a7F704D594d4d0C1B92Df34be",
		ApprovalAddress: "0x849B28659e5d612BDCfdA4aF89D00da3376400bD",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1: ",
		Description: sptypes.Description{
			Moniker:         "SP0",
			Identity:        "ID: SP0",
			Website:         "sp0.com",
			SecurityContact: "sp0_contact",
			Details:         "This is SP 0",
		},
	}); err != nil {
		log.Errorw("SetOwnSPInfo failed", "error", err)
		return err
	}
	log.Info("GetOwnSPInfo")
	ownSP, err := store.GetOwnSPInfo()
	if err != nil {
		log.Errorw("GetOwnSPInfo failed", "error", err)
	}
	log.Infow("GetOwnSPInfo", "ownSP", ownSP)
	return nil
}

func spInfoFunc(store *sqldb.SQLDB) error {
	log.Info("UpdateAllSP")
	if err := store.UpdateAllSP(spList()); err != nil {
		log.Errorw("UpdateAllSP failed", "error", err)
		return err
	}

	sps, err := store.FetchAllSP()
	if err != nil {
		log.Errorw("FetchAllSP failed", "error", err)
		return err
	}
	log.Infow("FetchAllSP", "sp list", sps)

	spFailed, err := store.FetchAllSP(sptypes.STATUS_IN_JAILED)
	if err != nil {
		log.Errorw("FetchAllSP failed", "error", err)
		return err
	}
	log.Infow("FetchAllSP in jail status", "sp failed", spFailed)

	sps1, err := store.FetchAllSPWithoutOwnSP()
	if err != nil {
		log.Errorw("FetchAllSPWithoutOwnSP failed", "error", err)
		return err
	}
	log.Infow("FetchAllSPWithoutOwnSP", "sp list", sps1)

	sps2, err := store.GetSPByAddress("0x3CA16ca2be371846511d153663F99584381ACc4D", sqldb.OperatorAddressType)
	if err != nil {
		log.Errorw("GetSPByAddress failed", "error", err)
		return err
	}
	log.Infow("GetSPByAddress", "sp list", sps2)

	sps3, err := store.GetSPByEndpoint("127.0.0.1:")
	if err != nil {
		log.Errorw("GetSPByEndpoint failed", "error", err)
		return err
	}
	log.Infow("GetSPByEndpoint", "sp list", sps3)

	return nil
}

func storageParamsFunc(store *sqldb.SQLDB) error {
	log.Info("SetStorageParams")
	if err := store.SetStorageParams(&storagetypes.Params{
		MaxSegmentSize:          16 * 1024 * 1024,
		RedundantDataChunkNum:   4,
		RedundantParityChunkNum: 2,
		MaxPayloadSize:          2 * 1024 * 1024 * 1024,
	}); err != nil {
		log.Errorw("SetStorageParams failed", "error", err)
		return err
	}

	log.Info("GetStorageParams")
	storageParams, err := store.GetStorageParams()
	if err != nil {
		log.Errorw("GetStorageParams failed", "error", err)
	}
	log.Infow("GetStorageParams", "storageParams", storageParams)
	return nil
}

func spList() []*sptypes.StorageProvider {
	sp0 := &sptypes.StorageProvider{
		OperatorAddress: "0xD30717BFD945901946Ec74f1986cca3fcf59681B",
		FundingAddress:  "0xbE2B3F7318f288E898E15553C6B9c15B5D1F3A1e",
		SealAddress:     "0x29735289CDbda24a7F704D594d4d0C1B92Df34be",
		ApprovalAddress: "0x849B28659e5d612BDCfdA4aF89D00da3376400bD",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://10.180.42.161:9033",
		Description: sptypes.Description{
			Moniker:         "SP0",
			Identity:        "ID: SP0",
			Website:         "sp0.com",
			SecurityContact: "sp0_contact",
			Details:         "This is SP 0",
		},
	}
	sp1 := &sptypes.StorageProvider{
		OperatorAddress: "0x3CA16ca2be371846511d153663F99584381ACc4D",
		FundingAddress:  "0x48CB0a6175a6f01afa4cA3D395cd21e9F1FE8b82",
		SealAddress:     "0xbB2fAdc9D63EcB4349679a2dA6Cf63A2b279Ee80",
		ApprovalAddress: "0x43aab10CdC8E97a7A659c02f8165ad853075bEfE",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1:9034",
		Description: sptypes.Description{
			Moniker:         "SP1",
			Identity:        "ID: SP1",
			Website:         "sp1.com",
			SecurityContact: "sp1_contact",
			Details:         "This is SP 1",
		},
	}
	sp2 := &sptypes.StorageProvider{
		OperatorAddress: "0xc46d9Fb40055FD896e02cB1c79B75d98d644a29a",
		FundingAddress:  "0x0A33851A1802C52f48D7B4d2A7cdF6b308f13D58",
		SealAddress:     "0xc7970e5107E62a643ad37C14560A3f7fD9C6bcC2",
		ApprovalAddress: "0x264C9121e58A482Ae4EFDd6b1163Bb00754038Ec",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1:9035",
		Description: sptypes.Description{
			Moniker:         "SP2",
			Identity:        "ID: SP2",
			Website:         "sp2.com",
			SecurityContact: "sp2_contact",
			Details:         "This is SP 2",
		},
	}
	sp3 := &sptypes.StorageProvider{
		OperatorAddress: "0x54D42b5F3a52360e2142D3fe3C25F789a6C8aeDe",
		FundingAddress:  "0x1fAFBc61aA8461b588730b5A7Fb0B61704Ac933B",
		SealAddress:     "0x9e1eC587692185393b5e90e95406eeaC3c543198",
		ApprovalAddress: "0x4c75C4157Afc03c9145DA95EB4513403fa055006",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1:9036",
		Description: sptypes.Description{
			Moniker:         "SP3",
			Identity:        "ID: SP3",
			Website:         "sp3.com",
			SecurityContact: "sp3_contact",
			Details:         "This is SP 3",
		},
	}
	sp4 := &sptypes.StorageProvider{
		OperatorAddress: "0x99837186527734EF3893567FDd44A25CE37d2c73",
		FundingAddress:  "0xe05Eb551442BC70BCF9e8857ffDd4c6d76d7B212",
		SealAddress:     "0x097f81f664300BF07FEe67Aee89F563b650d4f68",
		ApprovalAddress: "0xD0B6A4A1016673a8cBeBF03f277915CB1fBE0c43",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1:9037",
		Description: sptypes.Description{
			Moniker:         "SP4",
			Identity:        "ID: SP4",
			Website:         "sp4.com",
			SecurityContact: "sp4_contact",
			Details:         "This is SP 4",
		},
	}
	sp5 := &sptypes.StorageProvider{
		OperatorAddress: "0xd834B908f4a0a0DE10a438248C79e9442B87A5b6",
		FundingAddress:  "0xBEdD7602a816ec78b834C18ec959A44bB3C875dc",
		SealAddress:     "0xA04d3C9fa088de4F335e1ED0b2FEaa165C8B89c1",
		ApprovalAddress: "0x0Bc6576033F7e05cc7FBD10b9CDc284A44490078",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1:9038",
		Description: sptypes.Description{
			Moniker:         "SP5",
			Identity:        "ID: SP5",
			Website:         "sp5.com",
			SecurityContact: "sp5_contact",
			Details:         "This is SP 5",
		},
	}
	sp6 := &sptypes.StorageProvider{
		OperatorAddress: "0x05e69d497DbD18d2Ec7171821Df8C1ebc9F1550E",
		FundingAddress:  "0xa65B8d937ac6400Fc3cD91c1a333E55b02F1f1fc",
		SealAddress:     "0xfC8eBC2F58aF9140Aa2Df21A4499B838808D2ABa",
		ApprovalAddress: "0x1a0EaC20d38001fb6af6D2859C72359344deDB1b",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_SERVICE,
		Endpoint:        "http://127.0.0.1:9039",
		Description: sptypes.Description{
			Moniker:         "SP6",
			Identity:        "ID: SP6",
			Website:         "sp6.com",
			SecurityContact: "sp6_contact",
			Details:         "This is SP 6",
		},
	}
	spFailed := &sptypes.StorageProvider{
		OperatorAddress: "88888",
		FundingAddress:  "99999",
		SealAddress:     "11111",
		ApprovalAddress: "22222",
		TotalDeposit:    math.NewInt(10000000000000000000000),
		Status:          sptypes.STATUS_IN_JAILED,
		Endpoint:        "http://127.0.0.1: ",
		Description: sptypes.Description{
			Moniker:         "SP7",
			Identity:        "ID: SP7",
			Website:         "sp7.com",
			SecurityContact: "sp7_contact",
			Details:         "This is SP 7",
		},
	}
	return []*sptypes.StorageProvider{sp0, sp1, sp2, sp3, sp4, sp5, sp6, spFailed}
}
