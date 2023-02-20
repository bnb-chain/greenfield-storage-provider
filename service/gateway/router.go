package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := io.ReadAll(r.Body)
	log.Warnw("not found handler", "header", r.Header, "host", r.Host, "url", r.URL)
	w.WriteHeader(404)
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
		Queries(model.TransactionQuery, "").
		HandlerFunc(g.VerifySignature(g.putObjectTxHandler))
	//bucketRouter.NewRoute().
	//	Name("PutObject").
	//	Methods(http.MethodPut).
	//	Path("/{object:.+}").
	//	Queries(model.PutObjectV2Query, "").
	//	HandlerFunc(g.putObjectV2Handler)
	bucketRouter.NewRoute().
		Name("PutObject").
		Methods(http.MethodPut).
		Path("/{object:.+}").
		HandlerFunc(g.VerifySignature(g.putObjectV2Handler))
	bucketRouter.NewRoute().
		Name("CreateBucket").
		Methods(http.MethodPut).
		HandlerFunc(g.VerifySignature(g.createBucketHandler))
	bucketRouter.NewRoute().
		Name("GetObject").
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.VerifySignature(g.getObjectHandler))
	bucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	// todo: delete it in future
	_ = g.putObjectHandler
	// admin router, path style.
	/*
		adminRouter := r.PathPrefix(AdminPath).Subrouter()
		adminRouter.NewRoute().
			Name("GetApproval").
			Path(GetApprovalSubPath).
			Methods(http.MethodGet).
			Queries(ActionQuery, "{action}").
			HandlerFunc(g.getApprovalHandler)
		adminRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	*/
	r.Path(model.AdminPath+model.GetApprovalSubPath).
		Name("GetApproval").
		Methods(http.MethodGet).
		Queries(model.ActionQuery, "{action}").
		HandlerFunc(g.VerifySignature(g.getApprovalHandler))

	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
