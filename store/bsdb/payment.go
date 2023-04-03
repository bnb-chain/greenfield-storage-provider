package bsdb

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

// GetPaymentByBucketName get payment info by a bucket name
func (b *BsDBImpl) GetPaymentByBucketName(bucketName string, isFullList bool) (*StreamRecord, error) {
	var (
		streamRecord *StreamRecord
		err          error
		bucket       *Bucket
	)

	bucket, err = b.GetBucketByName(bucketName, isFullList)
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
func (b *BsDBImpl) GetPaymentByBucketID(bucketID int64, isFullList bool) (*StreamRecord, error) {
	var (
		streamRecord *StreamRecord
		err          error
		bucket       *Bucket
	)

	bucket, err = b.GetBucketByID(bucketID, isFullList)
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
