package gateway

import (
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

// echo impl
func (g *GatewayService) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := ioutil.ReadAll(r.Body)
	log.Warnw("not found handler", "header", r.Header, "host", r.Host)
	w.Write(s)
}

func (g *GatewayService) registerhandler(r *mux.Router) {
	bucketRouter := r.Host("{bucket:.+}." + g.config.Domain).Subrouter()
	bucketRouter.NewRoute().
		Name("PutObject").
		Methods(http.MethodPut).
		Path("/{object:.+}").
		Queries("transaction", "").
		HandlerFunc(g.putObjectTxHandler)
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
	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
