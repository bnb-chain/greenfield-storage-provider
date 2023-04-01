package stream

import (
	"fmt"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type payloadStreamTestContext struct {
	title            string // title of the test case
	initFunc         func() *PayloadStream
	sendFunc         func(*PayloadStream)
	entryNumber      int // expected entry number
	totalEntryLength int // expected sum of all the entry length
	lastEntryLength  int // expected the last entry length
	isError          bool
}

func checkPayloadStream(t *testing.T, ctx *payloadStreamTestContext) {
	var (
		err                  error
		realEntryNumber      int
		realTotalEntryLength int
		realLastEntryLength  int
	)
	s := ctx.initFunc()
	go ctx.sendFunc(s)

	for {
		entry, ok := <-s.AsyncStreamRead()
		if !ok { // has finished
			break
		}
		if entry.Error() != nil {
			err = entry.Error()
			break
		}
		log.Debugw("get piece entry from stream", "piece_key", entry.PieceKey(),
			"piece_len", len(entry.Data()), "error", entry.Error())
		realEntryNumber++
		realLastEntryLength = len(entry.Data())
		realTotalEntryLength += len(entry.Data())

	}

	if ctx.isError {
		require.Error(t, err)
	} else {
		assert.Equal(t, realEntryNumber, ctx.entryNumber)
		assert.Equal(t, realTotalEntryLength, ctx.totalEntryLength)
		assert.Equal(t, realLastEntryLength, ctx.lastEntryLength)
	}
}

func TestPayloadStream(t *testing.T) {
	testCases := []payloadStreamTestContext{
		{
			"invalid case",
			func() *PayloadStream {
				s := NewAsyncPayloadStream()
				_ = s.InitAsyncPayloadStream(1, storagetypes.REDUNDANCY_REPLICA_TYPE, 16*1024*1024, 1)
				return s
			},
			func(s *PayloadStream) {
				s.StreamWrite([]byte("s"))
				s.StreamCloseWithError(fmt.Errorf("invalid stream"))
			},
			1,
			1,
			1,
			true,
		},
		{
			"1 byte case",
			func() *PayloadStream {
				s := NewAsyncPayloadStream()
				_ = s.InitAsyncPayloadStream(1, storagetypes.REDUNDANCY_REPLICA_TYPE, 16*1024*1024, 1)
				return s
			},
			func(s *PayloadStream) {
				s.StreamWrite([]byte("s"))
				s.StreamClose()
			},
			1,
			1,
			1,
			false,
		},
		{
			"16MB byte case",
			func() *PayloadStream {
				s := NewAsyncPayloadStream()
				_ = s.InitAsyncPayloadStream(1, storagetypes.REDUNDANCY_REPLICA_TYPE, 16*1024*1024, 1)
				return s
			},
			func(s *PayloadStream) {
				s.StreamWrite(make([]byte, 16*1024*1024))
				s.StreamClose()
			},
			1,
			16 * 1024 * 1024,
			16 * 1024 * 1024,
			false,
		},
		{
			"16MB + 1 byte case",
			func() *PayloadStream {
				s := NewAsyncPayloadStream()
				_ = s.InitAsyncPayloadStream(1, storagetypes.REDUNDANCY_REPLICA_TYPE, 16*1024*1024, 1)
				return s
			},
			func(s *PayloadStream) {
				s.StreamWrite(make([]byte, 16*1024*1024+1))
				s.StreamClose()
			},
			2,
			16*1024*1024 + 1,
			1,
			false,
		},
		{
			"32MB byte case",
			func() *PayloadStream {
				s := NewAsyncPayloadStream()
				_ = s.InitAsyncPayloadStream(1, storagetypes.REDUNDANCY_REPLICA_TYPE, 16*1024*1024, 1)
				return s
			},
			func(s *PayloadStream) {
				s.StreamWrite(make([]byte, 32*1024*1024))
				s.StreamClose()
			},
			2,
			16 * 1024 * 1024 * 2,
			16 * 1024 * 1024,
			false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			checkPayloadStream(t, &testCase)
		})
	}
}
