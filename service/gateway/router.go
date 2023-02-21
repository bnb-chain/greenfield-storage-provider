package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	putObjectRouterName = "PutObject"
	getObjectRouterName = "GetObject"
	approvalRouterName  = "GetApproval"
	challengeRouterName = "Challenge"
)

const (
	createBucketApprovalAction = "CreateBucket"
	createObjectApprovalAction = "CreateObject"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := io.ReadAll(r.Body)
	log.Warnw("not found handler", "header", r.Header, "host", r.Host, "url", r.URL)
	w.WriteHeader(http.StatusNotFound)
	w.Write(s)
}

// registerHandler is used to register mux handlers.
func (g *Gateway) registerHandler(r *mux.Router) {
	// bucket router, virtual-hosted style
	bucketRouter := r.Host("{bucket:.+}." + g.config.Domain).Subrouter()
	bucketRouter.NewRoute().
		Name(putObjectRouterName).
		Methods(http.MethodPut).
		Path("/{object:.+}").
		HandlerFunc(g.putObjectHandler)
	bucketRouter.NewRoute().
		Name(getObjectRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	bucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

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
		Name(approvalRouterName).
		Methods(http.MethodGet).
		Queries(model.ActionQuery, "{action}").
		HandlerFunc(g.getApprovalHandler)
	// sync piece to syncer
	r.Path(model.SyncerPath).Name("SyncPiece").Methods(http.MethodPut).HandlerFunc(g.syncPieceHandler)

	r.Path(model.AdminPath + model.ChallengeSubPath).
		Name(challengeRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.challengeHandler)

	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
