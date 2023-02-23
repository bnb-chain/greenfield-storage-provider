package router

import (
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"github.com/gin-gonic/gin"
)

type Router struct {
	store store.IStore
}

func NewRouter(store store.IStore) *Router {
	return &Router{
		store: store,
	}
}

func (r *Router) InitHandlers(router *gin.RouterGroup) {
	//can use middleware here if needed in the future
	// router.Use(httputils.MetricsMiddleware)

	accounts := https.NewRouterGroup(router.Group("/accounts"))
	user := accounts.Group("/:account_id")
	user.GET("/buckets", https.Handler(r.GetUserBuckets))
	user.GET("/buckets/:bucket_name/objects", NewBucketNameWrapper(r.ListObjectsByBucketName))
}
