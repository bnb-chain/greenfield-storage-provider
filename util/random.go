package util

import (
	"math/rand"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	byteArr      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytesToLower = "abcdefghijklmnopqrstuvwxyz"
	bytesToUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func rdm(n int, data string) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = data[rand.Int63()%int64(len(data))]
	}
	return string(b)
}

func RandomNum(n, scope int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(n) + scope
}

// RandInt64 generate random int64 between min and max
func RandInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min)
}

func RandomString(n int) string {
	return rdm(n, byteArr)
}

func RandomStringToLower(n int) string {
	return rdm(n, bytesToLower)
}

func RandomStringToUpper(n int) string {
	return rdm(n, bytesToUpper)
}

func RandHexKey() string {
	key, _ := crypto.GenerateKey()

	keyBytes := crypto.FromECDSA(key)
	hexkey := hexutil.Encode(keyBytes)[2:]
	return hexkey
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// GetRandomObjectName generate random object name.
func GetRandomObjectName() string {
	return randString(10)
}

// GetRandomBucketName generate random bucket name.
func GetRandomBucketName() string {
	return randString(5)
}

// GetRandomGroupName generate random group name.
func GetRandomGroupName() string {
	return randString(7)
}
