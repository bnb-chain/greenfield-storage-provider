package https

import "github.com/gin-gonic/gin"

type RouterGroup struct {
	ginRG *gin.RouterGroup
}

func NewRouterGroup(ginGR *gin.RouterGroup) *RouterGroup {
	return &RouterGroup{
		ginRG: ginGR,
	}
}

func (rg *RouterGroup) GET(relativePath string, handler IHandler) {
	rg.ginRG.GET(relativePath, toGinHandler(handler))
}

func (rg *RouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouterGroup {
	return &RouterGroup{rg.ginRG.Group(relativePath, handlers...)}
}
