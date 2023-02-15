package router

import (
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
)

const (
	CTXHandlerKey  = "_CTX_Handler"
	CtxResponseKey = "_CTX_RESPONSE"
)

const (
	CodeOk = 20000

	ErrorCodeBadRequest = 401
	ErrorCodeNotFound   = 404

	ErrorCodeInternalError = 501
)

type Router struct {
	store store.IStore
}

func NewRouter(store store.IStore) *Router {
	return &Router{
		store: store,
	}
}

type RouterGroup struct {
	ginRG *gin.RouterGroup
}

func NewRouterGroup(ginGR *gin.RouterGroup) *RouterGroup {
	return &RouterGroup{
		ginRG: ginGR,
	}
}

type IHandler interface {
	Handler() Handler
	MethodName() string
}

type Handler func(*gin.Context) (resp interface{}, err *Error)

type Error struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Errors  []SubError `json:"errors,omitempty"`
}

type SubError struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// https://google.github.io/styleguide/jsoncstyleguide.xml#JSON_Structure_&_Reserved_Property_Names
type Response struct {
	Error *Error      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func toGinHandler(h IHandler) gin.HandlerFunc {
	methodName := h.MethodName()
	f := h.Handler()
	return func(ctx *gin.Context) {
		resp, err := f(ctx)
		hresp := &Response{
			Error: err,
			Data:  resp,
		}

		ctx.JSON(http.StatusOK, hresp)
		ctx.Set(CTXHandlerKey, methodName)
		ctx.Set(CtxResponseKey, hresp)
	}
}

func (rg *RouterGroup) GET(relativePath string, handler IHandler) {
	rg.ginRG.GET(relativePath, toGinHandler(handler))
}

func (rg *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{rg.ginRG.Group(relativePath, handlers...)}
}

func (h Handler) Handler() Handler {
	return h
}

func (h Handler) MethodName() string {
	return filepath.Base(getFunctionName(h))
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (r *Router) InitHandlers(router *gin.RouterGroup) {
	//can use middleware here if needed in the future
	// router.Use(httputils.MetricsMiddleware)

	accounts := NewRouterGroup(router.Group("/accounts"))
	user := accounts.Group("/:account_id")
	user.GET("/buckets", Handler(r.GetUserBuckets))

}
