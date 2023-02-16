package https

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
)

const (
	CTXHandlerKey = "_CTX_Handler"
)

type IHandler interface {
	Handler() Handler
	MethodName() string
}

type Handler func(*gin.Context) (resp interface{}, err *Error)

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

func (h Handler) Handler() Handler {
	return h
}

func (h Handler) MethodName() string {
	return filepath.Base(getFunctionName(h))
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
