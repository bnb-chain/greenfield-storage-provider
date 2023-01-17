package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

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

// case1 128bytes, Inline type, do not need to be segmented(< segment size, 16MB).
func runCase1() {
	log.Info("start run case1(128byte, Inline type)")
	// get auth
	{
		log.Infow("start get auth")
		url := "http://" + gatewayAddress + "/greenfield/admin/v1/get-approval?action=createObject"
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get auth failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.BFSResourceHeader, "test_bucket/case1")
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get auth failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get auth failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish get auth",
			"preSign", res.Header.Get("X-Bfs-Pre-Signature"),
			"statusCode", res.StatusCode)
	}
	// put object
	{
		log.Infow("start prepare data for put object")
		buf := generateRandString(64)
		log.Infow("finish prepare data for put object")

		log.Infow("start put object")
		url := "http://" + gatewayAddress + "/case1?putobjectv2"
		method := "PUT"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = "test_bucket.bfs.nodereal.com"
		req.Header.Add(model.BFSTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "1")
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("put object failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
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
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = "test_bucket.bfs.nodereal.com"
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
	// get auth
	{
		log.Infow("start get auth")
		url := "http://" + gatewayAddress + "/greenfield/admin/v1/get-approval?action=createObject"
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get auth failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.BFSResourceHeader, "test_bucket/case2")
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get auth failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get auth failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish get auth",
			"preSign", res.Header.Get("X-Bfs-Pre-Signature"),
			"statusCode", res.StatusCode)
	}
	// put object
	{
		log.Infow("start prepare data for put object")
		buf := generateRandString(64 * 1024 * 1024)
		log.Infow("finish prepare data for put object")

		log.Infow("start put object")
		url := "http://" + gatewayAddress + "/case2?putobjectv2"
		method := "PUT"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = "test_bucket.bfs.nodereal.com"
		req.Header.Add(model.BFSTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "67108864")
		req.Header.Add(model.BFSRedundancyTypeHeader, model.ReplicaRedundancyTypeHeaderValue)
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("put object failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
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
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = "test_bucket.bfs.nodereal.com"
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
	// get auth
	{
		log.Infow("start get auth")
		url := "http://" + gatewayAddress + "/greenfield/admin/v1/get-approval?action=createObject"
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get auth failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.BFSResourceHeader, "test_bucket/case3")
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get auth failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get auth failed, due to read response body", "error", err)
			return
		}
		log.Infow("finish get auth",
			"preSign", res.Header.Get("X-Bfs-Pre-Signature"),
			"statusCode", res.StatusCode)
	}
	// put object
	{
		log.Infow("start prepare data for put object")
		buf := generateRandString(200 * 1024 * 1024)
		log.Infow("finish prepare data for put object")

		log.Infow("start put object")
		url := "http://" + gatewayAddress + "/case3?putobjectv2"
		method := "PUT"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = "test_bucket.bfs.nodereal.com"
		req.Header.Add(model.BFSTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "209715200")
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("put object failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)
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
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = "test_bucket.bfs.nodereal.com"
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
