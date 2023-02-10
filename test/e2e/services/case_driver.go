package main

import (
	"bytes"
	"flag"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-sdk-go/pkg/signer"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
<<<<<<< HEAD

=======
>>>>>>> e160384 (feat: v 0.0.2 version  (#88))
	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	configFile     = flag.String("config", "./config.toml", "config file path")
	letters        = []byte("0123456789")
	gatewayAddress string
)

func generateRandString(n int) string {
	rand.Seed(time.Now().Unix())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

const (
	getMethod  = "GET"
	putMethod  = "PUT"
	testDomain = "test_bucket.gnfd.nodereal.com"
)

func mockSignRequest(request *http.Request) error {
	privKey, _, _ := testdata.KeyEthSecp256k1TestPubAddr()
	err := signer.SignRequest(request, privKey, signer.AuthInfo{
		SignType:        model.SignTypeV1,
		MetaMaskSignStr: "",
	})
	if err != nil {
		log.Errorw("mock signature failed, due to ", "error", err)
		return err
	}
	return nil
}

// case1 128bytes, Inline type, do not need to be segmented(< segment size, 16MB).
func runCase1() {
	log.Info("start run case1(128byte, Inline type)")
	// get approval
	{
		log.Infow("start get approval")
		url := "http://" + gatewayAddress + "/greenfield/admin/v1/get-approval?action=createObject"
		method := getMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get approval failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdResourceHeader, "test_bucket/case1")
		if err = mockSignRequest(req); err != nil {
			log.Errorw("get approval failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get approval failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get approval failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish get approval",
			"preSign", res.Header.Get("X-Bfs-Pre-Signature"),
			"statusCode", res.StatusCode)
	}
	// put object
	{
		log.Infow("start prepare data for put object")
		buf := generateRandString(64)
		log.Infow("finish prepare data for put object")

		log.Infow("start put object")
		url := "http://" + gatewayAddress + "/case1"
		method := putMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = testDomain
		req.Header.Add(model.GnfdTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "64")
		if err = mockSignRequest(req); err != nil {
			log.Errorw("put object failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("put object failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("put object failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish put object",
			"etag", res.Header.Get(model.ETagHeader),
			"statusCode", res.StatusCode)
	}
	// get object
	{
		log.Infow("start get object")
		url := "http://" + gatewayAddress + "/case1"
		method := getMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = testDomain
		if err = mockSignRequest(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))
	}
	log.Info("end run case1")
}

// case2 64MB, Replica type, should be segmented.
func runCase2() {
	log.Info("start run case2(64MB, Replica type)")
	// get approval
	{
		log.Infow("start get approval")
		url := "http://" + gatewayAddress + "/greenfield/admin/v1/get-approval?action=createObject"
		method := getMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get approval failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdResourceHeader, "test_bucket/case2")
		if err = mockSignRequest(req); err != nil {
			log.Errorw("get approval failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get approval failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get approval failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish get approval",
			"preSign", res.Header.Get("X-Bfs-Pre-Signature"),
			"statusCode", res.StatusCode)
	}
	// put object
	{
		log.Infow("start prepare data for put object")
		buf := generateRandString(64 * 1024 * 1024)
		log.Infow("finish prepare data for put object")

		log.Infow("start put object")
		url := "http://" + gatewayAddress + "/case2"
		method := putMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = testDomain
		req.Header.Add(model.GnfdTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "67108864")
		req.Header.Add(model.GnfdRedundancyTypeHeader, model.ReplicaRedundancyTypeHeaderValue)
		if err = mockSignRequest(req); err != nil {
			log.Errorw("put object failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("put object failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("put object failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish put object",
			"etag", res.Header.Get(model.ETagHeader),
			"statusCode", res.StatusCode)
	}
	// get object
	{
		log.Infow("start get object")
		url := "http://" + gatewayAddress + "/case2"
		method := getMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = testDomain
		if err = mockSignRequest(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))

	}
	log.Info("end run case2")
}

// case3 200MB, EC type, should be segmented.
func runCase3() {
	log.Info("start run case3(200MB, EC type)")
	// get approval
	{
		log.Infow("start get approval")
		url := "http://" + gatewayAddress + "/greenfield/admin/v1/get-approval?action=createObject"
		method := getMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get approval failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdResourceHeader, "test_bucket/case3")
		if err = mockSignRequest(req); err != nil {
			log.Errorw("get approval failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get approval failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get approval failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish get approval",
			"preSign", res.Header.Get("X-Bfs-Pre-Signature"),
			"statusCode", res.StatusCode)
	}
	// put object
	{
		log.Infow("start prepare data for put object")
		buf := generateRandString(200 * 1024 * 1024)
		log.Infow("finish prepare data for put object")

		log.Infow("start put object")
		url := "http://" + gatewayAddress + "/case3"
		method := putMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = testDomain
		req.Header.Add(model.GnfdTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "209715200")
		if err = mockSignRequest(req); err != nil {
			log.Errorw("put object failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("put object failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("put object failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish put object",
			"etag", res.Header.Get(model.ETagHeader),
			"statusCode", res.StatusCode)
	}
	// get object
	{
		log.Infow("start get object")
		url := "http://" + gatewayAddress + "/case3"
		method := getMethod
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = testDomain
		if err = mockSignRequest(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))

	}
	log.Info("end run case3")
}

func main() {
	log.Info("start run cases")

	rand.Seed(time.Now().Unix())
	cfg := config.LoadConfig(*configFile)
	gatewayAddress = cfg.GatewayCfg.Address

	runCase1()
	runCase2()
	runCase3()

	log.Info("end run cases")
}
