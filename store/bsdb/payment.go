package bsdb

//
//import (
//	"errors"
//
//	"github.com/forbole/juno/v4/common"
//	"gorm.io/gorm"
//)
//
//// GetPaymentByBucketName get payment info by a bucket name
//func (b *BsDBImpl) GetPaymentByBucketName(bucketName string, includePrivate bool) (*StreamRecord, error) {
//	var (
//		streamRecord *StreamRecord
//		err          error
//		bucket       *Bucket
//	)
//
//	bucket, err = b.GetBucketByName(bucketName, includePrivate)
//	if err != nil {
//		return nil, err
//	}
//
//	if bucket == nil {
//		return nil, nil
//	}
//
//	streamRecord, err = b.GetPaymentByPaymentAddress(bucket.PaymentAddress)
//	return streamRecord, err
//}
//
//// GetPaymentByBucketID get payment info by a bucket id
//func (b *BsDBImpl) GetPaymentByBucketID(bucketID int64, includePrivate bool) (*StreamRecord, error) {
//	var (
//		streamRecord *StreamRecord
//		err          error
//		bucket       *Bucket
//	)
//
//	bucket, err = b.GetBucketByID(bucketID, includePrivate)
//	if err != nil {
//		return nil, err
//	}
//
//	if bucket == nil {
//		return nil, nil
//	}
//
//	streamRecord, err = b.GetPaymentByPaymentAddress(bucket.PaymentAddress)
//	return streamRecord, err
//}
//
//// GetPaymentByPaymentAddress get payment info by a payment address
//func (b *BsDBImpl) GetPaymentByPaymentAddress(paymentAddress common.Address) (*StreamRecord, error) {
//	var (
//		streamRecord *StreamRecord
//		err          error
//	)
//
//	err = b.db.Table((&StreamRecord{}).TableName()).Take(&streamRecord, "account = ?", paymentAddress).Error
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		return nil, nil
//	}
//
//	return streamRecord, err
//}
