package storage

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateChecksum(t *testing.T) {
	r := bytes.NewReader([]byte("test"))
	assert.NotNil(t, r)
	cases := []struct {
		name         string
		rs           io.ReadSeeker
		wantedResult string
	}{
		{
			name:         "bytes Reader",
			rs:           r,
			wantedResult: "2258662080",
		},
		{
			name:         "strings Reader",
			rs:           strings.NewReader("hello world"),
			wantedResult: "3381945770",
		},
		{
			name:         "returned checksum is empty",
			rs:           strings.NewReader(""),
			wantedResult: "0",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := generateChecksum(tt.rs)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func Test_checksumReaderRead(t *testing.T) {
	cases := []struct {
		name      string
		cr        *checksumReader
		buf       []byte
		wantedN   int
		wantedErr error
	}{
		{
			name: "1",
			cr: &checksumReader{
				ReadCloser: io.NopCloser(strings.NewReader("hello world")),
			},
			buf:       []byte("test"),
			wantedN:   4,
			wantedErr: nil,
		},
		{
			name: "2",
			cr: &checksumReader{
				ReadCloser: io.NopCloser(strings.NewReader("")),
				expected:   5,
			},
			buf:       []byte("test"),
			wantedN:   0,
			wantedErr: errors.New("failed to verify checksum: 0 != 5"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.cr.Read(tt.buf)
			assert.Equal(t, tt.wantedN, result)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func Test_verifyChecksum(t *testing.T) {
	cases := []struct {
		name         string
		rc           io.ReadCloser
		checksum     string
		wantedResult io.ReadCloser
	}{
		{
			name:         "1",
			rc:           io.NopCloser(strings.NewReader("hello world")),
			checksum:     "",
			wantedResult: io.NopCloser(strings.NewReader("hello world")),
		},
		{
			name:         "2",
			rc:           io.NopCloser(strings.NewReader("hello world")),
			checksum:     "abc",
			wantedResult: io.NopCloser(strings.NewReader("hello world")),
		},
		{
			name:     "3",
			rc:       io.NopCloser(strings.NewReader("hello world")),
			checksum: "10",
			wantedResult: &checksumReader{
				ReadCloser: io.NopCloser(strings.NewReader("hello world")),
				expected:   10,
				checksum:   0,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyChecksum(tt.rc, tt.checksum)
			assert.Equal(t, result, tt.wantedResult)
		})
	}
}
