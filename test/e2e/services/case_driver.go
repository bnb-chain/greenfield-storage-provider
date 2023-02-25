package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/keys"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	// configFile is gateway config, case_driver parse gateway related config for connecting sp
	configFile = flag.String("config", "./config.toml", "gateway config file path")

	// gateway real ip, the domain name can be configured when there is a real gateway domain name
	gatewayAddress string

	// testBucketName is a testcase bucket name
	testBucketName = "sp_test_bucket"

	// hostHeader is virtual hosted style, include bucket_name and gateway domain_name
	hostHeader string

	metadataAddress string
)

func generateRandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = []byte("0123456789")[rand.Intn(len([]byte("0123456789")))]
	}
	return string(b)
}

func generateRequestSignature(request *http.Request) error {
	privKey, _, _ := testdata.KeyEthSecp256k1TestPubAddr()
	keyManager, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	if err != nil {
		log.Errorw("failed to new private key manager")
	}
	client, err := sp.NewSpClientWithKeyManager("gnfd.nodereal.com", &sp.Option{}, keyManager)
	if err != nil {
		log.Errorw("failed to new SP client with key manager")
	}
	err = client.SignRequest(request, sp.NewAuthInfo(false, ""))
	if err != nil {
		log.Errorw("failed to mock signature", "error", err)
		return err
	}
	return nil
}

func checkIntegrityHash(integrityHash string, pieceHashList string, index int, payload []byte) error {
	h, err := hex.DecodeString(integrityHash)
	if err != nil {
		return err
	}
	hashList, err := util.DecodePieceHash(pieceHashList)
	if err != nil {
		return err
	}
	return hash.CheckIntegrityHash(h, hashList, index, payload)
}

// case1 128bytes, Inline type, do not need to be segmented(< segment size, 16MB).
func runCase1() {
	//var objectID uint64
	objectName := "case1_object_name"
	log.Info("start run case1(128byte, Inline type)")
	// get approval
	{
		log.Infow("start get approval")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+model.AdminPath+model.GetApprovalSubPath+"?action=createObject",
			strings.NewReader(""))
		if err != nil {
			log.Errorw("get approval failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdResourceHeader, testBucketName+"/"+objectName)
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get approval failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
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
		req, err := http.NewRequest(http.MethodPut,
			"http://"+gatewayAddress+"/"+objectName,
			strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		req.Header.Add(model.GnfdTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "64")
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("put object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
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
		_, err = util.HeaderToUint64(res.Header.Get(model.GnfdObjectIDHeader))
		if err != nil {
			log.Errorw("put object failed, due to has no object id", "error", err)
			return
		}
	}
	// get object
	//{
	//	log.Infow("start get object")
	//	req, err := http.NewRequest(http.MethodGet,
	//		"http://"+gatewayAddress+"/"+objectName,
	//		strings.NewReader(""))
	//	if err != nil {
	//		log.Errorw("get object failed, due to new request", "error", err)
	//		return
	//	}
	//	req.Host = hostHeader
	//	if err = generateRequestSignature(req); err != nil {
	//		log.Errorw("get object failed, due to sign signature", "error", err)
	//		return
	//	}
	//	client := &http.Client{}
	//	res, err := client.Do(req)
	//	if err != nil {
	//		log.Errorw("get object failed, due to send request", "error", err)
	//		return
	//	}
	//	buf := new(bytes.Buffer)
	//	buf.ReadFrom(res.Body)
	//	log.Infow("finish get object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))
	//}
	//// get range object
	//{
	//	log.Infow("start get range object")
	//	req, err := http.NewRequest(http.MethodGet,
	//		"http://"+gatewayAddress+"/"+objectName,
	//		strings.NewReader(""))
	//	if err != nil {
	//		log.Errorw("get object failed, due to new request", "error", err)
	//		return
	//	}
	//	req.Host = hostHeader
	//	req.Header.Add(model.RangeHeader, "bytes=1-")
	//
	//	if err = generateRequestSignature(req); err != nil {
	//		log.Errorw("get object failed, due to sign signature", "error", err)
	//		return
	//	}
	//	client := &http.Client{}
	//	res, err := client.Do(req)
	//	if err != nil {
	//		log.Errorw("get object failed, due to send request", "error", err)
	//		return
	//	}
	//	buf := new(bytes.Buffer)
	//	buf.ReadFrom(res.Body)
	//	log.Infow("finish get range object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))
	//}
	//// wait update meta
	//time.Sleep(5 * time.Second)
	//// challenge piece
	//{
	//	log.Infow("start challenge piece")
	//	req, err := http.NewRequest(http.MethodGet,
	//		"http://"+gatewayAddress+model.AdminPath+model.ChallengeSubPath,
	//		strings.NewReader(""))
	//	if err != nil {
	//		log.Errorw("challenge failed, due to new request", "error", err)
	//		return
	//	}
	//	req.Header.Add(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
	//	req.Header.Add(model.GnfdPieceIndexHeader, "0")
	//	req.Header.Add(model.GnfdRedundancyIndexHeader, "0")
	//	if err = generateRequestSignature(req); err != nil {
	//		log.Errorw("challenge failed, due to sign signature", "error", err)
	//		return
	//	}
	//	client := &http.Client{}
	//	res, err := client.Do(req)
	//	if err != nil {
	//		log.Errorw("challenge failed, due to send request", "error", err)
	//		return
	//	}
	//	defer res.Body.Close()
	//	buf, err := io.ReadAll(res.Body)
	//	if err != nil {
	//		log.Errorw("challenge failed, due to read response body", "error", err)
	//		return
	//	}
	//	err = checkIntegrityHash(res.Header.Get(model.GnfdIntegrityHashHeader), res.Header.Get(model.GnfdPieceHashHeader), 0, buf)
	//	if err != nil {
	//		log.Errorw("challenge failed, due to checkIntegrityHash", "error", err)
	//		return
	//	}
	//	log.Infow("finish challenge", "statusCode", res.StatusCode)
	//}
	log.Info("end run case1")
}

// case2 64MB, Replica type, should be segmented.
func runCase2() {
	var objectID uint64
	objectName := "case2_object_name"
	log.Info("start run case2(64MB, Replica type)")
	// get approval
	{
		log.Infow("start get approval")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+model.AdminPath+model.GetApprovalSubPath+"?action=createObject",
			strings.NewReader(""))
		if err != nil {
			log.Errorw("get approval failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdResourceHeader, testBucketName+"/"+objectName)
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get approval failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
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
		req, err := http.NewRequest(http.MethodPut,
			"http://"+gatewayAddress+"/"+objectName,
			strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		req.Header.Add(model.GnfdTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "67108864")
		req.Header.Add(model.GnfdRedundancyTypeHeader, model.ReplicaRedundancyTypeHeaderValue)
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("put object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
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
		objectID, err = util.HeaderToUint64(res.Header.Get(model.GnfdObjectIDHeader))
		if err != nil {
			log.Errorw("put object failed, due to has no object id", "error", err)
			return
		}
	}
	// get object
	{
		log.Infow("start get object")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+"/"+objectName,
			strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))

	}
	// get range object
	{
		log.Infow("start get range object")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+"/"+objectName,
			strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		req.Header.Add(model.RangeHeader, "bytes=17825792-35651583") // 17MB, and in two segment

		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get range object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))
	}
	// wait update meta
	time.Sleep(5 * time.Second)
	// challenge piece
	{
		log.Infow("start challenge piece")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+model.AdminPath+model.ChallengeSubPath,
			strings.NewReader(""))
		if err != nil {
			log.Errorw("challenge failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
		req.Header.Add(model.GnfdPieceIndexHeader, "1")
		req.Header.Add(model.GnfdRedundancyIndexHeader, "0")
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("challenge failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("challenge failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("challenge failed, due to read response body", "error", err)
			return
		}
		err = checkIntegrityHash(res.Header.Get(model.GnfdIntegrityHashHeader), res.Header.Get(model.GnfdPieceHashHeader), 1, buf)
		if err != nil {
			log.Errorw("challenge failed, due to checkIntegrityHash", "error", err)
			return
		}
		log.Infow("finish challenge", "statusCode", res.StatusCode)
	}
	log.Info("end run case2")
}

// case3 200MB, EC type, should be segmented.
func runCase3() {
	var objectID uint64
	objectName := "case3_object_name"
	log.Info("start run case3(200MB, EC type)")
	// get approval
	{
		log.Infow("start get approval")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+model.AdminPath+model.GetApprovalSubPath+"?action=createObject",
			strings.NewReader(""))
		if err != nil {
			log.Errorw("get approval failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdResourceHeader, testBucketName+"/"+objectName)
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get approval failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
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
		req, err := http.NewRequest(http.MethodPut, "http://"+gatewayAddress+"/"+objectName, strings.NewReader(buf))
		if err != nil {
			log.Errorw("put object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		req.Header.Add(model.GnfdTransactionHashHeader, generateRandString(64))
		req.Header.Add(model.ContentLengthHeader, "209715200")
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("put object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
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
		objectID, err = util.HeaderToUint64(res.Header.Get(model.GnfdObjectIDHeader))
		if err != nil {
			log.Errorw("put object failed, due to has no object id", "error", err)
			return
		}
	}
	// get object
	{
		log.Infow("start get object")
		req, err := http.NewRequest(http.MethodGet, "http://"+gatewayAddress+"/"+objectName, strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))
	}
	// get range object
	{
		log.Infow("start get range object")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+"/"+objectName,
			strings.NewReader(""))
		if err != nil {
			log.Errorw("get object failed, due to new request", "error", err)
			return
		}
		req.Host = hostHeader
		req.Header.Add(model.RangeHeader, "bytes=35651584-") // 266MB

		if err = generateRequestSignature(req); err != nil {
			log.Errorw("get object failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get object failed, due to send request", "error", err)
			return
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		log.Infow("finish get range object payload", "statusCode", res.StatusCode, "body len", len(buf.String()))
	}
	// wait update meta
	time.Sleep(5 * time.Second)
	// challenge piece
	{
		log.Infow("start challenge piece")
		req, err := http.NewRequest(http.MethodGet,
			"http://"+gatewayAddress+model.AdminPath+model.ChallengeSubPath,
			strings.NewReader(""))
		if err != nil {
			log.Errorw("challenge failed, due to new request", "error", err)
			return
		}
		req.Header.Add(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
		req.Header.Add(model.GnfdPieceIndexHeader, "10")
		req.Header.Add(model.GnfdRedundancyIndexHeader, "0")
		if err = generateRequestSignature(req); err != nil {
			log.Errorw("challenge failed, due to sign signature", "error", err)
			return
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("challenge failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("challenge failed, due to read response body", "error", err)
			return
		}
		err = checkIntegrityHash(res.Header.Get(model.GnfdIntegrityHashHeader), res.Header.Get(model.GnfdPieceHashHeader), 10, buf)
		if err != nil {
			log.Errorw("challenge failed, due to checkIntegrityHash", "error", err)
			return
		}
		log.Infow("finish challenge", "statusCode", res.StatusCode)
	}
	log.Info("end run case3")
}

// case4
func runCase4() {
	log.Info("start run case4(GetUserBuckets)")
	//get user buckets
	{
		log.Infow("start get user buckets")
		account := "46765cbc-d30c-4f4a-a814-b68181fcab12"
		url := fmt.Sprintf("http://%s/accounts/%s/buckets", metadataAddress, account)
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get user buckets failed, due to new request", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get user buckets failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get user buckets failed, due to read response body", "error", err)
			return
		}

		log.Infow("finish get user buckets", "statusCode", res.StatusCode)
	}
	log.Info("end run case4")
}

// case5
func runCase5() {
	log.Info("start run case5(ListObjectsByBucketName)")
	// list objects by bucket name
	{
		log.Infow("start get user objects")
		account := "46765cbc-d30c-4f4a-a814-b68181fcab12"
		bucketName := "hello world"
		url := fmt.Sprintf("http://%s/accounts/%s/buckets/%s/objects", metadataAddress, account, bucketName)
		method := "GET"
		client := &http.Client{}
		req, err := http.NewRequest(method, url, strings.NewReader(""))
		if err != nil {
			log.Errorw("get user objects failed, due to new request", "error", err)
			return
		}
		res, err := client.Do(req)
		if err != nil {
			log.Errorw("get user objects failed, due to send request", "error", err)
			return
		}
		defer res.Body.Close()
		_, err = io.ReadAll(res.Body)
		if err != nil {
			log.Errorw("get user objects failed, due to read response body", "error", err)
			return
		}

		log.Infow("finish get user objects", "statusCode", res.StatusCode)
	}
	log.Info("end run case5")
}

func main() {
	log.Info("start run cases")

	rand.Seed(time.Now().Unix())
	cfg := config.LoadConfig(*configFile)
	gatewayAddress = cfg.GatewayCfg.Address
	hostHeader = testBucketName + "." + cfg.GatewayCfg.Domain
	//metadataAddress = cfg.MetadataCfg.Address

	runCase1()
	//runCase2()
	//runCase3()
	//runCase4()
	//runCase5()

	log.Info("end run cases")
}
