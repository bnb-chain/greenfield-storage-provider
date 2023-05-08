package gfsperrors

import (
	"net/http"
	"sort"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func (m *GfSpError) Error() string {
	return m.String()
}

var code2Err *GfSpErrorMap
var once sync.Once

func init() {
	once.Do(func() {
		code2Err = &GfSpErrorMap{
			innerCode2Err: make(map[int32]*GfSpError),
		}
	})
}

func Register(codeSpace string, httpStatuesCode int, innerCode int, description string) *GfSpError {
	err := &GfSpError{
		CodeSpace:      codeSpace,
		HttpStatusCode: int32(httpStatuesCode),
		InnerCode:      int32(innerCode),
		Description:    description,
	}
	code2Err.AddErr(err)
	return err
}

func GfSpErrorList() []*GfSpError {
	var list []*GfSpError
	for _, err := range code2Err.innerCode2Err {
		list = append(list, err)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].InnerCode < list[j].InnerCode
	})
	return list
}

const (
	DefaultCodeSpace = "GfSp"
	DefaultInnerCode = 999999
)

func MakeGfSpError(err error) *GfSpError {
	switch e := err.(type) {
	case *GfSpError:
		return e
	default:
		return &GfSpError{
			CodeSpace:      DefaultCodeSpace,
			HttpStatusCode: int32(http.StatusInternalServerError),
			InnerCode:      int32(DefaultInnerCode),
			Description:    err.Error(),
		}
	}
}

type GfSpErrorMap struct {
	innerCode2Err map[int32]*GfSpError
	mux           sync.RWMutex
}

func (g *GfSpErrorMap) AddErr(err *GfSpError) {
	g.mux.Lock()
	defer g.mux.Unlock()
	if old, ok := g.innerCode2Err[err.InnerCode]; ok {
		log.Panicf("[%s] and [%s] error code conflict !!!", err.Error(), old.Error())
	}
	g.innerCode2Err[err.InnerCode] = err
}

func (g *GfSpErrorMap) GfSpErr(innerCode int) *GfSpError {
	g.mux.RLock()
	defer g.mux.RUnlock()
	err, ok := g.innerCode2Err[int32(innerCode)]
	if !ok {
		return nil
	}
	return err
}
