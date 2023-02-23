package router

import (
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"github.com/gin-gonic/gin"
)

func (r *Router) GetUserBuckets(ctx *gin.Context) (resp interface{}, herr *https.Error) {
	buckets, err := r.store.GetUserBuckets(ctx)
	if err != nil {
		log.Infof("err:%+v", err)
		return nil, https.NewInternalError("fail to get user buckets, the error is: %s", err.Error())
	}

	return buckets, nil
}
