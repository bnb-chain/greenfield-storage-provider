package gater

import (
	"testing"

	"github.com/gorilla/mux"
)

var (
	testDomain = "www.route-test.com"
	gw         = &GateModular{
		domain: testDomain,
	}
	scheme     = "https://"
	bucketName = "test-bucket-name"
	objectName = "test-object-name"
)

func setupRouter(t *testing.T) *mux.Router {
	gwRouter := mux.NewRouter().SkipClean(true)
	gw.RegisterHandler(gwRouter)
	return gwRouter
}
