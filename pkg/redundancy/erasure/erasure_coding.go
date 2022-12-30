package erasure

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/klauspost/reedsolomon"
)

// RSEncoder - reedSolomon RSEncoder encoding details.
type RSEncoder struct {
	encoder                  func() reedsolomon.Encoder
	dataShards, parityShards int
	blockSize                int64 // the data size to be encoded
}

// NewRSEncoder creates a new RSEncoder with reedsolomon encoder
func NewRSEncoder(dataShards, parityShards int, blockSize int64) (r RSEncoder, err error) {
	// Check the parameters for sanity now.
	if dataShards <= 0 || parityShards < 0 {
		return r, reedsolomon.ErrInvShardNum
	}

	if dataShards+parityShards > 256 {
		return r, reedsolomon.ErrMaxShardNum
	}

	r = RSEncoder{
		dataShards:   dataShards,
		parityShards: parityShards,
		blockSize:    blockSize,
	}

	var encoder reedsolomon.Encoder
	var once sync.Once
	r.encoder = func() reedsolomon.Encoder {
		once.Do(func() {
			r, err := reedsolomon.New(dataShards, parityShards,
				reedsolomon.WithAutoGoroutines(int(r.ShardSize())))

			if err != nil {
				log.Printf("new RS encoder fail, dataShards %d, parityShards %d", dataShards,
					parityShards)
			}
			encoder = r
		})
		return encoder
	}

	return
}

// EncodeData encodes the given data and returns the reedsolomon encoded shards
func (r *RSEncoder) EncodeData(content []byte) ([][]byte, error) {
	if len(content) == 0 {
		return make([][]byte, r.dataShards+r.parityShards), nil
	}
	encoded, err := r.encoder().Split(content)
	if err != nil {
		log.Println("encoder split data error ", err.Error())
		return nil, err
	}
	if err = r.encoder().Encode(encoded); err != nil {
		log.Println("encoder encode error ", err.Error())
		return nil, err
	}
	return encoded, nil
}

// DecodeDataShards decodes the input erasure encoded data shards data.
// The func will recreate any missing data shards if possible.
func (r *RSEncoder) DecodeDataShards(content [][]byte) error {
	emptyShardNum := 0
	for _, b := range content {
		if len(b) == 0 {
			emptyShardNum++
			continue
		}
	}
	if emptyShardNum == len(content) {
		return nil
	}
	return r.encoder().ReconstructData(content)
}

// DecodeShards decodes the input erasure encoded data and verifies it.
// The func recreate the missing shards if possible.
func (r *RSEncoder) DecodeShards(data [][]byte) error {
	if err := r.encoder().Reconstruct(data); err != nil {
		log.Println("recreate the missing shard fail", err)
		return err
	}
	ok, err := r.encoder().Verify(data)
	if err != nil {
		log.Println("decode verify fail", err.Error())
		return err
	}

	if !ok {
		return fmt.Errorf("parity shards contained incorrect data.")
	}
	return nil
}

// ShardSize - returns actual shared size from blockSize.
func (r *RSEncoder) ShardSize() int64 {
	shardNum := int64(r.dataShards)
	n := r.blockSize / shardNum
	if r.blockSize > 0 && r.blockSize%shardNum != 0 {
		n++
	}
	return n
}

// GetOriginalData decode the shards and reconstruct the original content
func (r *RSEncoder) GetOriginalData(shardsData [][]byte, originLength int64) ([]byte, error) {
	err := r.DecodeDataShards(shardsData)
	if err != nil {
		log.Printf("decode shards fail:%s", err.Error())
		return []byte(""), err
	}

	var buffer bytes.Buffer
	for i := 0; i < r.dataShards; i++ {
		buffer.Write(shardsData[i])
	}
	shardSize := r.ShardSize()
	deCodeBytes := buffer.Bytes()
	// ignore padding content
	if shardSize*int64(r.dataShards) >= originLength {
		deCodeBytes = deCodeBytes[:originLength]
	}

	return deCodeBytes, nil
}
