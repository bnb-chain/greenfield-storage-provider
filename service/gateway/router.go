package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	putObjectRouterName = "PutObject"
	getObjectRouterName = "GetObject"
	approvalRouterName  = "GetApproval"
	challengeRouterName = "Challenge"
	syncPieceRouterName = "SyncPiece"
)

const (
	createBucketApprovalAction = "CreateBucket"
	createObjectApprovalAction = "CreateObject"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Errorw("not found handler", "header", r.Header, "host", r.Host, "url", r.URL)
	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorw("failed to read the unknown request", "error", err)
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write(s)
}

// registerHandler is used to register mux handlers.
func (g *Gateway) registerHandler(r *mux.Router) {
	// bucket router, virtual-hosted style
	bucketRouter := r.Host("{bucket:.+}." + g.config.Domain).Subrouter()
	//bucketRouter.NewRoute().
	//	Name(putObjectRouterName).
	//	Methods(http.MethodPut).
	//	Path("/{object:.+}").
	//	HandlerFunc(g.putObjectHandler)
	//bucketRouter.NewRoute().
	//	Name(getObjectRouterName).
	//	Methods(http.MethodGet).
	//	Path("/{object:.+}").
	//	HandlerFunc(g.getObjectHandler)
	bucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	// admin router, path style.
	//r.Path(model.AdminPath+model.GetApprovalSubPath).
	//	Name(approvalRouterName).
	//	Methods(http.MethodGet).
	//	Queries(model.ActionQuery, "{action}").
	//	HandlerFunc(g.getApprovalHandler)
	//r.Path(model.AdminPath + model.ChallengeSubPath).
	//	Name(challengeRouterName).
	//	Methods(http.MethodGet).
	//	HandlerFunc(g.challengeHandler)
	// sync piece to syncer
	//r.Path(model.SyncerPath).
	//	Name(syncPieceRouterName).
	//	Methods(http.MethodPut).
	//	HandlerFunc(g.syncPieceHandler)

	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
