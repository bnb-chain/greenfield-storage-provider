package stopserving

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	lru "github.com/hashicorp/golang-lru"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metadataclient "github.com/bnb-chain/greenfield-storage-provider/service/metadata/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	signerclient "github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
)

var _ lifecycle.Service = &StopServing{}

var (
	// DiscontinueReason defines the reason for stop serving
	DiscontinueReason = "testnet cleanup"

	// FetchBucketsInterval define the interval to fetch buckets for stop serving
	FetchBucketsInterval = 5 * time.Minute
	// SubmitTransactionInterval defines the interval between two stop serving transaction submissions
	SubmitTransactionInterval = 500 * time.Millisecond
)

// StopServing module is responsible for stop serving buckets on testnet.
type StopServing struct {
	config   *StopServingConfig
	cache    *lru.Cache
	running  atomic.Value
	stopCh   chan struct{}
	signer   *signerclient.SignerClient
	metadata *metadataclient.MetadataClient
}

// NewStopServingService returns an instance of stop serving
func NewStopServingService(cfg *StopServingConfig) (*StopServing, error) {
	var (
		stopServing *StopServing
		err         error
	)

	stopServing = &StopServing{
		config: cfg,
		stopCh: make(chan struct{}),
	}
	if stopServing.cache, err = lru.New(model.LruCacheLimit); err != nil {
		log.Errorw("failed to create lru cache", "error", err)
		return nil, err
	}
	if stopServing.signer, err = signerclient.NewSignerClient(cfg.SignerGrpcAddress); err != nil {
		log.Errorw("failed to create signer client", "error", err)
		return nil, err
	}
	if stopServing.metadata, err = metadataclient.NewMetadataClient(cfg.MetadataGrpcAddress); err != nil {
		log.Errorw("failed to create metadata client", "error", err)
		return nil, err
	}
	log.Debugw("stop serving service created successfully")
	return stopServing, nil
}

// Name return the stop serving service name
func (s *StopServing) Name() string {
	return model.StopServingService
}

// Start function start background goroutine to execute stop serving
func (s *StopServing) Start(ctx context.Context) error {
	if s.running.Swap(true) == true {
		return errors.New("stop serving has already started")
	}

	if s.config.DiscontinueConfig.BucketKeepAliveDays > 0 {
		// start background task
		go s.eventLoop()
	}

	return nil
}

// eventLoop a background goroutine to periodically conduct stop serving
func (s *StopServing) eventLoop() {
	s.discontinueBuckets()
	ticker := time.NewTicker(FetchBucketsInterval)
	for {
		select {
		case <-ticker.C:
			s.discontinueBuckets()
		case <-s.stopCh:
			return
		}
	}
}

// discontinueBuckets fetch buckets from metadata service and submit transactions to chain
func (s *StopServing) discontinueBuckets() {
	//TODO: replace with ListExpiredBucketsBySp after metadata is ready.
	//TODO: parameters 1) sp address 2) creation time
	req := &types.GetUserBucketsRequest{AccountId: s.config.SpOperatorAddress}
	resp, err := s.metadata.GetUserBuckets(context.Background(), req)
	if err != nil {
		log.Errorw("failed to query bucket info", "error", err)
		return
	}

	for _, bucket := range resp.Buckets {
		bucketName := bucket.BucketInfo.BucketName
		if s.cache.Contains(bucketName) { // this bucket has been discontinued, however the metadata indexer did not handle it yet
			continue
		}

		discontinueBucket := &storagetypes.MsgDiscontinueBucket{
			BucketName: bucket.BucketInfo.BucketName,
			Reason:     DiscontinueReason,
		}
		txHash, err := s.signer.DiscontinueBucketOnChain(context.Background(), discontinueBucket)
		if err != nil {
			log.Errorw("failed to discontinue bucket on chain", "error", err)
			return
		} else {
			s.cache.Add(bucket.BucketInfo.BucketName, struct{}{})
			log.Infow("succeed to discontinue bucket", "bucket_name", discontinueBucket.BucketName, "tx_hash", txHash)
		}
		time.Sleep(SubmitTransactionInterval)
	}
}

// Stop stops serving background goroutine
func (s *StopServing) Stop(ctx context.Context) error {
	if s.running.Swap(false) == false {
		return errors.New("stop serving has already stop")
	}
	close(s.stopCh)
	return nil
}
