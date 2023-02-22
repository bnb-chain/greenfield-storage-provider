package sdk

import (
	"net/http"
)

// IHTTPClient is an interface generated for "net/http.Client".
type IHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
	Get(string) (*http.Response, error)
}
