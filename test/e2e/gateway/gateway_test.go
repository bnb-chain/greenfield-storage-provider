package e2e

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
)

var (
	g       *gateway.Gateway
	payload string
	md5str  string
)

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateRandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func setUp() {
	c := gateway.DefaultGatewayConfig
	c.Address = "127.0.0.1:9099"
	c.UploaderConfig.DebugDir = "./tmptestdata"
	c.DownloaderConfig.DebugDir = "./tmptestdata"
	c.ChainConfig.DebugDir = "./tmptestdata"

	g, _ = gateway.NewGatewayService(c)
	_ = g.Start(context.Background())
	payload = generateRandString(65 * 1024)
	w := md5.New()
	io.WriteString(w, payload)
	md5str = fmt.Sprintf("%x", w.Sum(nil))
}

func tearDown() {
	os.RemoveAll("./tmptestdata")
	// g.Stop(context.Background())
}

const tesDomain = "test.bfs.nodereal.com"

func TestGateway(t *testing.T) {
	setUp()
	// create bucket
	{
		// succeed
		req1, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/", nil)
		req1.Host = tesDomain
		res1, _ := http.DefaultClient.Do(req1)
		assert.Equal(t, 200, res1.StatusCode)

		// failed due to has existed
		req2, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/", nil)
		req2.Host = tesDomain
		res2, _ := http.DefaultClient.Do(req2)
		assert.Equal(t, 409, res2.StatusCode)
	}
	// put object
	{
		// failed due to bucket not found
		req1, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject", strings.NewReader(payload))
		req1.Host = tesDomain
		res1, _ := http.DefaultClient.Do(req1)
		assert.Equal(t, 500, res1.StatusCode)

		// failed due to tx not found
		req2, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject", strings.NewReader(payload))
		req2.Host = tesDomain
		res2, _ := http.DefaultClient.Do(req2)
		assert.Equal(t, 404, res2.StatusCode)

		// succeed
		req3, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject?transaction", nil)
		req3.Host = tesDomain
		res3, _ := http.DefaultClient.Do(req3)
		assert.Equal(t, 200, res3.StatusCode)

		req4, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject", strings.NewReader(payload))
		req4.Host = tesDomain
		res4, _ := http.DefaultClient.Do(req4)
		assert.Equal(t, 200, res4.StatusCode)
		assert.Equal(t, md5str, res4.Header.Get("etag"))
	}
	// get object
	{
		// failed due to object not found
		req1, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:9099/testobjecta", nil)
		req1.Host = tesDomain
		res1, _ := http.DefaultClient.Do(req1)
		assert.Equal(t, 404, res1.StatusCode)
		// succeed
		req2, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:9099/testobject", nil)
		req2.Host = tesDomain
		res2, _ := http.DefaultClient.Do(req2)
		assert.Equal(t, 200, res2.StatusCode)
		// assert.Equal(t, res2.Header.Get("Content-Length"), 65*1024)
	}
	tearDown()
}
