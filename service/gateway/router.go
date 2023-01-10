package gateway

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := ioutil.ReadAll(r.Body)
	log.Warnw("not found handler", "header", r.Header, "host", r.Host, "url", r.URL)
	w.Write(s)

}

// registerhandler is used to register mux handlers.
func (g *Gateway) registerhandler(r *mux.Router) {
	// bucket router, virtual-hosted style
	bucketRouter := r.Host("{bucket:.+}." + g.config.Domain).Subrouter()
	bucketRouter.NewRoute().
		Name("PutObject").
		Methods(http.MethodPut).
		Path("/{object:.+}").
		Queries(TransactionQuery, "").
		HandlerFunc(g.putObjectTxHandler)
	bucketRouter.NewRoute().
		Name("PutObject").
		Methods(http.MethodPut).
		Path("/{object:.+}").
		Queries(PutObjectV2Query, "").
		HandlerFunc(g.putObjectV2Handler)
	bucketRouter.NewRoute().
		Name("PutObject").
		Methods(http.MethodPut).
		Path("/{object:.+}").
		HandlerFunc(g.putObjectHandler)
	bucketRouter.NewRoute().
		Name("CreateBucket").
		Methods(http.MethodPut).
		HandlerFunc(g.createBucketHandler)
	bucketRouter.NewRoute().
		Name("GetObject").
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	bucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	// admin router, path style.
	adminRouter := r.PathPrefix(AdminPath).Subrouter()
	adminRouter.NewRoute().
		Name("GetAuthentication").
		Path(GetApprovalSubPath).
		Methods(http.MethodGet).
		Queries(ActionQuery, "{action}").
		HandlerFunc(g.getAuthenticationHandler)
	adminRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
