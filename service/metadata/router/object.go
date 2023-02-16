package router

import (
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"github.com/gin-gonic/gin"
)

func NewBucketNameWrapper(f func(ctx *gin.Context, bucketName string) (resp interface{}, herr *https.Error)) https.Handler {
	return func(c *gin.Context) (resp interface{}, err *https.Error) {
		bucketName := c.Param("bucket_name")
		if bucketName == "" {
			return nil, &https.Error{
				Code:    https.ErrorCodeInternalError,
				Message: "bucket name cannot be empty",
			}
		}

		return f(c, bucketName)
	}
}

func (r *Router) ListObjectsByBucket(ctx *gin.Context, bucketName string) (resp interface{}, herr *https.Error) {
	buckets, err := r.store.ListObjectsByBucketName(ctx, bucketName)
	if err != nil {
		log.Infof("err:%+v", err)
		return nil, https.NewInternalError("fail to list objects by bucket name", err)
	}

	return buckets, nil
}
