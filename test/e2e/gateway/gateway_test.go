package e2e

import (
	"crypto/md5"
	"fmt"
	"github.com/bnb-chain/inscription-storage-provider/service/gateway"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
)

var (
	g       *gateway.GatewayService
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
	g = gateway.NewGatewayService()
	g.Init("./testdata/gateway.toml")
	g.Start()
	payload = generateRandString(65 * 1024)

	w := md5.New()
	io.WriteString(w, payload)
	md5str = fmt.Sprintf("%x", w.Sum(nil))
}

func tearDown() {
	g.Stop()
	os.RemoveAll("./tmptestdata")
}

func TestGateway(t *testing.T) {
	setUp()
	// create bucket
	{
		// succeed
		req1, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/", nil)
		req1.Host = "test.bfs.nodereal.com"
		res1, _ := http.DefaultClient.Do(req1)
		assert.Equal(t, res1.StatusCode, 200)

		// failed due to has existed
		req2, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/", nil)
		req2.Host = "test.bfs.nodereal.com"
		res2, _ := http.DefaultClient.Do(req2)
		assert.Equal(t, res2.StatusCode, 409)
	}
	// put object
	{
		// failed due to bucket not found
		req1, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject", strings.NewReader(payload))
		req1.Host = "testa.bfs.nodereal.com"
		res1, _ := http.DefaultClient.Do(req1)
		assert.Equal(t, res1.StatusCode, 500)

		// failed due to tx not found
		req2, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject", strings.NewReader(payload))
		req2.Host = "test.bfs.nodereal.com"
		res2, _ := http.DefaultClient.Do(req2)
		assert.Equal(t, res2.StatusCode, 404)

		// succeed
		req3, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject?transaction", nil)
		req3.Host = "test.bfs.nodereal.com"
		res3, _ := http.DefaultClient.Do(req3)
		assert.Equal(t, res3.StatusCode, 200)

		req4, _ := http.NewRequest(http.MethodPut, "http://127.0.0.1:9099/testobject", strings.NewReader(payload))
		req4.Host = "test.bfs.nodereal.com"
		res4, _ := http.DefaultClient.Do(req4)
		assert.Equal(t, res4.StatusCode, 200)
		assert.Equal(t, res4.Header.Get("etag"), md5str)
	}
	// get object
	{
		// failed due to object not found
		req1, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:9099/testobjecta", nil)
		req1.Host = "test.bfs.nodereal.com"
		res1, _ := http.DefaultClient.Do(req1)
		assert.Equal(t, res1.StatusCode, 404)
		// succeed
		req2, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:9099/testobject", nil)
		req2.Host = "test.bfs.nodereal.com"
		res2, _ := http.DefaultClient.Do(req2)
		assert.Equal(t, res2.StatusCode, 200)
		// assert.Equal(t, res2.Header.Get("Content-Length"), 65*1024)
	}
	tearDown()
}
