package client

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	protoHTTP  = "http"
	protoHTTPS = "https"
	protoWSS   = "wss"
	protoWS    = "ws"
	protoTCP   = "tcp"
	protoUNIX  = "unix"
)

// -------------------------------------------------------------

// Parsed URL structure
type parsedURL struct {
	url.URL

	isUnixSocket bool
}

// Parse URL and set defaults
func newParsedURL(remoteAddr string) (*parsedURL, error) {
	u, err := url.Parse(remoteAddr)
	if err != nil {
		return nil, err
	}

	// default to tcp if nothing specified
	if u.Scheme == "" {
		u.Scheme = protoTCP
	}

	pu := &parsedURL{
		URL:          *u,
		isUnixSocket: false,
	}

	if u.Scheme == protoUNIX {
		pu.isUnixSocket = true
	}

	return pu, nil
}

// SetDefaultSchemeHTTP Change protocol to HTTP for unknown protocols and TCP protocol - useful for RPC connections
func (u *parsedURL) SetDefaultSchemeHTTP() {
	// protocol to use for http operations, to support both http and https
	switch u.Scheme {
	case protoHTTP, protoHTTPS, protoWS, protoWSS:
		// known protocols not changed
	default:
		// default to http for unknown protocols (ex. tcp)
		u.Scheme = protoHTTP
	}
}

// GetHostWithPath Get full address without the protocol - useful for Dialer connections
func (u parsedURL) GetHostWithPath() string {
	// Remove protocol, userinfo and # fragment, assume opaque is empty
	return u.Host + u.EscapedPath()
}

// GetDialAddress returns the endpoint to dial for the parsed URL
func (u parsedURL) GetDialAddress() string {
	// if it's not a unix socket we return the host, example: localhost:443
	if !u.isUnixSocket {
		return u.Host
	}
	// otherwise we return the path of the unix socket, ex /tmp/socket
	return u.GetHostWithPath()
}

// -------------------------------------------------------------

func makeDialContext(dialer *net.Dialer, remoteAddr string) (func(ctx context.Context, network, address string) (net.Conn, error), error) {
	u, err := newParsedURL(remoteAddr)
	if err != nil {
		return nil, err
	}

	protocol := u.Scheme

	// accept http(s) as an alias for tcp
	switch protocol {
	case protoHTTP, protoHTTPS:
		protocol = protoTCP
	}

	dialFn := func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialer.DialContext(ctx, protocol, u.GetDialAddress())
	}

	return dialFn, nil
}

// DefaultHTTPClient is used to create a http client with some default parameters.
// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://).
// An error will be returned in case of invalid remoteAddr.
func DefaultHTTPClient(remoteAddr string) (*http.Client, error) {
	dialer := &net.Dialer{
		Timeout:   time.Second,
		KeepAlive: 60 * time.Second,
	}

	dailContext, err := makeDialContext(dialer, remoteAddr)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		// Set to true to prevent GZIP-bomb DoS attacks
		DisableCompression:  true,
		DialContext:         dailContext,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     1000,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		// Please care about this timeout, it's the timeout for all request
		Timeout:   10 * time.Minute,
		Transport: transport,
	}

	return client, nil
}
