package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"io"
	"net/http"
	"time"
)

type Client struct {
	name    string
	httpCli IHTTPClient
	option  *Option
}

type Option struct {
	HTTPTimeout time.Duration
	Address     string
}

func (c *Client) get(ctx context.Context, url string, data interface{}, options ...httpOption) (err error) {
	return c.do(ctx, http.MethodGet, url, data, options...)
}

type httpCallOption struct {
	// for case that url contains uuid,
	// we should use a custom name for label "handler"
	metricsHandlerName string
	body               io.Reader
}

func (c *httpCallOption) ApplyOptions(options ...httpOption) {
	for _, opt := range options {
		opt(c)
	}
}

type httpOption func(*httpCallOption)

func withCustomMetricsHandlerName(handler string) httpOption {
	return func(c *httpCallOption) {
		c.metricsHandlerName = handler
	}
}

func (c *Client) do(ctx context.Context, method string, url string, data interface{}, options ...httpOption) (err error) {
	opt := &httpCallOption{}
	opt.ApplyOptions(options...)

	defer func(t0 time.Time) {
		handler := opt.metricsHandlerName
		if handler == "" {
			handler = url
		}

	}(time.Now())

	req, err := http.NewRequestWithContext(ctx, method, url, opt.body)
	if err != nil {
		return err
	}

	req.Header.Add(https.HTTPHeaderFrom, c.name)
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %+v", resp)
	}

	var hresp https.Response
	hresp.Data = data
	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bts, &hresp); err != nil {
		return err
	}

	if hresp.Error != nil {
		return hresp.Error
	}

	return nil
}
