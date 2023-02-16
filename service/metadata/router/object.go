package router

import (
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"github.com/gin-gonic/gin"
)

func NewBucketNameWrapper(f func(ctx *gin.Context, bucketName string) (resp interface{}, herr *Error)) Handler {
	return func(c *gin.Context) (resp interface{}, err *Error) {
		bucketName := c.Param("bucket_name")
		if bucketName == "" {
			return nil, &Error{
				Code:    ErrorCodeInternalError,
				Message: "bucket name cannot be empty",
			}
		}

		return f(c, bucketName)
	}
}

func (r *Router) ListObjectsByBucket(ctx *gin.Context, bucketName string) (resp interface{}, herr *Error) {
	buckets, err := r.store.ListObjectsByBucketName(ctx, bucketName)
	if err != nil {
		log.Infof("err:%+v", err)
		return nil, NewInternalError("fail to list objects by bucket name", err)
	}

	return buckets, nil
}
