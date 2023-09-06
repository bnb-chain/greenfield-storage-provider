package gater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasInvalidPath(t *testing.T) {
	cases := []struct {
		name         string
		path         string
		wantedResult bool
	}{
		{
			name:         "path contains .",
			path:         "https://test.com/./test-path",
			wantedResult: true,
		},
		{
			name:         "path contains ..",
			path:         "https://test.com/../test-path",
			wantedResult: true,
		},
		{
			name:         "path doesn't contains . and ..",
			path:         "https://test.com/test-path",
			wantedResult: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := hasInvalidPath(tt.path)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func Test_checkValidObjectPrefix(t *testing.T) {
	cases := []struct {
		name         string
		prefix       string
		wantedResult bool
	}{
		{
			name:         "invalid path",
			prefix:       "https://test.com/./test-path",
			wantedResult: false,
		},
		{
			name:         "invalid utf8 string",
			prefix:       string([]byte{0xff, 0xfe, 0xfd}),
			wantedResult: false,
		},
		{
			name:         "prefix contains //",
			prefix:       "//testObjectName",
			wantedResult: false,
		},
		{
			name:         "valid prefix",
			prefix:       "testObjectName",
			wantedResult: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := checkValidObjectPrefix(tt.prefix)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}
