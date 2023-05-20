package gfsperrors

import (
	"net/http"
	"sort"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	// DefaultCodeSpace defines the default code space for not predefined error.
	DefaultCodeSpace = "GfSp"
	// DefaultInnerCode defines the default inner code for not predefined error.
	DefaultInnerCode = 999999
)

// Error implements the error interface for compatibility with built-in error.
func (m *GfSpError) Error() string {
	return m.String()
}

// SetError sets the Description field by Error(), it is use for change the
// predefined Description.
// Example: fill the predefined error's Description according to client info.
func (m *GfSpError) SetError(err error) {
	if err == nil {
		return
	}
	m.Description = err.Error()
}

var (
	// gfspErrManager defines the Global GfSpError manager for managing the
	// predefined GfSpError.
	gfspErrManager *GfSpErrorManager
	once           sync.Once
)

func init() {
	once.Do(func() {
		gfspErrManager = &GfSpErrorManager{
			innerCode2Err: make(map[int32]*GfSpError),
		}
	})
}

// MakeGfSpError returns an GfSpError from the build-in error interface. It is
// difficult to predefine all errors. For undefined errors, there needs to be a
// way to capture them and return them to the client according to the GfSpError
// format specified by the system. Of course, this is a guaranteed solution, the
// error should be well-defined is the most ideal.
//
// If the input is not the GfSpError type, will use the DefaultCodeSpace and
// DefaultInnerCode, they are predefined by the reversed value, do not conflict
// with other predefined errors.
func MakeGfSpError(err error) *GfSpError {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *GfSpError:
		if e.GetInnerCode() == 0 {
			return nil
		}
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

// Register the predefined an error to the global error manager. Every error
// that needs to be displayed to the client needs to meet the format of GfSpError
// and be pre-defined in advance.
func Register(codeSpace string, httpStatuesCode int, innerCode int, description string) *GfSpError {
	err := &GfSpError{
		CodeSpace:      codeSpace,
		HttpStatusCode: int32(httpStatuesCode),
		InnerCode:      int32(innerCode),
		Description:    description,
	}
	gfspErrManager.AddErr(err)
	return err
}

// GfSpErrorList returns the list of predefined errors, it is useful for query
// all predefined errors detailed info.
// Example:
//
//	list the errors by cli... etc. tools under troubleshooting.
func GfSpErrorList() []*GfSpError {
	var list []*GfSpError
	for _, err := range gfspErrManager.innerCode2Err {
		list = append(list, err)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].InnerCode < list[j].InnerCode
	})
	return list
}

// GfSpErrorManager manages the predefined GfSpError, the GfSpError uses as
// the standard error format inside the SP system, it includes the CodeSpace,
// HttpStatusCode, InnerCode and Description fields. The HttpStatusCode uses
// to fill the http response header, the InnerCode and Description uses to
// fill the http response body if the request is failed, the CodeSpace and
// InnerCode uses to help developer tp quickly pinpoint the cause and location
// of errors.
//
// The InnerCode detailed specifications can be found at:
//
//	"github.com/bnb-chain/greenfield-storage-provider/base/errors.md"
type GfSpErrorManager struct {
	innerCode2Err map[int32]*GfSpError
	mux           sync.RWMutex
}

// AddErr add an error to the manager, predefined errors need to ensure the
// uniqueness of inner error codes.
func (g *GfSpErrorManager) AddErr(err *GfSpError) {
	g.mux.Lock()
	defer g.mux.Unlock()
	if old, ok := g.innerCode2Err[err.InnerCode]; ok {
		log.Panicf("[%s] and [%s] error code conflict !!!", err.Error(), old.Error())
	}
	g.innerCode2Err[err.InnerCode] = err
}
