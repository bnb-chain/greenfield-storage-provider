package client

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newParsedURL(t *testing.T) {
	cases := []struct {
		name         string
		remoteAddr   string
		wantedScheme string
		wantedIsErr  bool
	}{
		{
			name:         "http url",
			remoteAddr:   "http://127.0.0.1:9000",
			wantedScheme: protoHTTP,
			wantedIsErr:  false,
		},
		{
			name:         "no scheme",
			remoteAddr:   "john%20doe@www.google.com/",
			wantedScheme: protoTCP,
			wantedIsErr:  false,
		},
		{
			name:         "unix url",
			remoteAddr:   "unix://webmaster@www.google.com/",
			wantedScheme: protoUNIX,
			wantedIsErr:  false,
		},
		{
			name:         "wrong url",
			remoteAddr:   "\n\t",
			wantedScheme: "",
			wantedIsErr:  true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newParsedURL(tt.remoteAddr)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedScheme, result.Scheme)
			}
		})
	}
}

func TestSetDefaultSchemeHTTP(t *testing.T) {
	wsURL, err := newParsedURL("ws://www.example.com/socketserver")
	assert.Nil(t, err)
	ftpURL, err := newParsedURL("ftp://www.example.com/socketserver")
	assert.Nil(t, err)
	cases := []struct {
		name         string
		parsedURL    *parsedURL
		wantedResult string
	}{
		{
			name:         "ws scheme",
			parsedURL:    wsURL,
			wantedResult: protoWS,
		},
		{
			name:         "http scheme",
			parsedURL:    ftpURL,
			wantedResult: protoHTTP,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.parsedURL.SetDefaultSchemeHTTP()
			assert.Equal(t, tt.wantedResult, tt.parsedURL.Scheme)
		})
	}
}

func TestGetHostWithPath(t *testing.T) {
	wsURL, err := newParsedURL("ws://www.example.com/socketserver")
	assert.Nil(t, err)
	result := wsURL.GetHostWithPath()
	assert.Equal(t, "www.example.com/socketserver", result)
}

func TestGetDialAddress(t *testing.T) {
	wsURL, err := newParsedURL("ws://www.example.com/socketserver")
	assert.Nil(t, err)
	unixURL, err := newParsedURL("unix://a/b/c")
	assert.Nil(t, err)
	cases := []struct {
		name         string
		parsedURL    *parsedURL
		wantedResult string
	}{
		{
			name:         "ws url",
			parsedURL:    wsURL,
			wantedResult: "www.example.com",
		},
		{
			name:         "unix url",
			parsedURL:    unixURL,
			wantedResult: "a/b/c",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parsedURL.GetDialAddress()
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func Test_makeDialContext(t *testing.T) {
	cases := []struct {
		name        string
		dialer      *net.Dialer
		remoteAddr  string
		wantedIsErr bool
	}{
		{
			name: "http url",
			dialer: &net.Dialer{
				Timeout:   time.Second,
				KeepAlive: 60 * time.Second,
			},
			remoteAddr:  "http://127.0.0.1:9000",
			wantedIsErr: false,
		},
		{
			name:        "wrong remote address",
			dialer:      nil,
			remoteAddr:  "\n\t",
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := makeDialContext(tt.dialer, tt.remoteAddr)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
				_, _ = result(context.TODO(), "tcp", "localhost:0")
			}
		})
	}
}

func TestDefaultHTTPClient(t *testing.T) {
	cases := []struct {
		name         string
		remoteAddr   string
		wantedIsErr  bool
		wantedResult time.Duration
	}{
		{
			name:         "Set default http client successfully",
			remoteAddr:   "http://127.0.0.1:9000",
			wantedIsErr:  false,
			wantedResult: 10 * time.Minute,
		},
		{
			name:        "Set default http client unsuccessfully",
			remoteAddr:  "\n\t",
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DefaultHTTPClient(tt.remoteAddr)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantedResult, result.Timeout)
			}
		})
	}
}
