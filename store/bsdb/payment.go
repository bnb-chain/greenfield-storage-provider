package bsdb

import (
	"errors"
	"time"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetPaymentByBucketName get payment info by a bucket name
func (b *BsDBImpl) GetPaymentByBucketName(bucketName string, includePrivate bool) (*StreamRecord, error) {
	var (
		streamRecord *StreamRecord
		err          error
		bucket       *Bucket
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	bucket, err = b.GetBucketByName(bucketName, includePrivate)
	if err != nil {
		return nil, err
	}

	if bucket == nil {
		return nil, nil
	}

	streamRecord, err = b.GetPaymentByPaymentAddress(bucket.PaymentAddress)
	return streamRecord, err
}

// GetPaymentByBucketID get payment info by a bucket id
func (b *BsDBImpl) GetPaymentByBucketID(bucketID int64, includePrivate bool) (*StreamRecord, error) {
	var (
		streamRecord *StreamRecord
		err          error
		bucket       *Bucket
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	bucket, err = b.GetBucketByID(bucketID, includePrivate)
	if err != nil {
		return nil, err
	}

	if bucket == nil {
		return nil, nil
	}

	streamRecord, err = b.GetPaymentByPaymentAddress(bucket.PaymentAddress)
	return streamRecord, err
}

// GetPaymentByPaymentAddress get payment info by a payment address
func (b *BsDBImpl) GetPaymentByPaymentAddress(paymentAddress common.Address) (*StreamRecord, error) {
	var (
		streamRecord *StreamRecord
		err          error
	)

	err = b.db.Table((&StreamRecord{}).TableName()).Take(&streamRecord, "account = ?", paymentAddress).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return streamRecord, err
}

// ListUserPaymentAccounts list payment accounts by owner address
func (b *BsDBImpl) ListUserPaymentAccounts(accountID common.Address) ([]*StreamRecordPaymentAccount, error) {
	var (
		payments []*StreamRecordPaymentAccount
		err      error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()
	err = b.db.Table((&PaymentAccount{}).TableName()).
		Select("*").
		Joins("LEFT JOIN stream_records ON stream_records.account = payment_accounts.addr").
		Where(`payment_accounts.owner = ?`, accountID).
		Order("stream_records.account").
		Find(&payments).Error

	return payments, err
}

// ListPaymentAccountStreams list payment account streams
func (b *BsDBImpl) ListPaymentAccountStreams(paymentAccount common.Address) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()
	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where(`payment_address = ? and removed = false`, paymentAccount).
		Order("bucket_id").
		Find(&buckets).Error

	return buckets, err
}
