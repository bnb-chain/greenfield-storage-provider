package storage

import (
	"compress/flate"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mockMemoryBucket = "memoryBucket"

func setupMemoryTest(t *testing.T) *memoryStore {
	return &memoryStore{name: mockMemoryBucket}
}

func TestMemoryStore_String(t *testing.T) {
	store := setupMemoryTest(t)
	result := store.String()
	assert.Equal(t, "memory://memoryBucket/", result)
}

func TestMemoryStore_GetObjectSuccess(t *testing.T) {
	cases := []struct {
		name         string
		key          string
		offset       int64
		limit        int64
		wantedResult string
		wantedErr    error
	}{
		{
			name:         "memory_get_success_test1",
			key:          mockKey,
			offset:       0,
			limit:        0,
			wantedResult: mockAccessKey,
			wantedErr:    nil,
		},
		{
			name:         "memory_get_success_test2",
			key:          mockKey,
			offset:       15,
			limit:        5,
			wantedResult: "",
			wantedErr:    nil,
		},
		{
			name:         "memory_get_success_test3",
			key:          mockKey,
			offset:       10,
			limit:        1,
			wantedResult: "K",
			wantedErr:    nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			store.objects = map[string]*memoryObject{
				mockKey: {data: []byte(mockAccessKey)},
			}
			data, err := store.GetObject(context.TODO(), tt.key, tt.offset, tt.limit)
			assert.Equal(t, tt.wantedErr, err)
			data1, err := io.ReadAll(data)
			if err != nil {
				t.Fatalf("io ReadAll error: %s", err)
			}
			assert.Equal(t, tt.wantedResult, string(data1))
		})
	}
}

func TestMemoryStore_GetObjectFailure(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		wantedErr error
	}{
		{
			name:      "memory_get_error_test1",
			key:       emptyString,
			wantedErr: ErrInvalidObjectKey,
		},
		{
			name:      "memory_get_error_test2",
			key:       mockKey,
			wantedErr: ErrNoSuchObject,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			store.objects = map[string]*memoryObject{
				mockSecretKey: {data: []byte(mockSecretKey)},
			}
			data, err := store.GetObject(context.TODO(), tt.key, 0, 0)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, nil, data)
		})
	}
}

func TestMemoryStore_PutObject(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		reader    io.Reader
		wantedErr error
	}{
		{
			name:      "memory_put_test1",
			key:       emptyString,
			reader:    strings.NewReader(mockEndpoint),
			wantedErr: ErrInvalidObjectKey,
		},
		{
			name:      "memory_put_test2",
			key:       mockAccessKey,
			reader:    strings.NewReader(mockEndpoint),
			wantedErr: nil,
		},
		{
			name:      "memory_put_test3",
			key:       mockAccessKey,
			reader:    flate.NewReader(strings.NewReader("test")),
			wantedErr: io.ErrUnexpectedEOF,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			store.objects = map[string]*memoryObject{
				mockAccessKey: {data: []byte(mockSecretKey)},
			}
			err := store.PutObject(context.TODO(), tt.key, tt.reader)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestMemoryStore_DeleteObject(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		wantedErr error
	}{
		{
			name:      "memory_delete_test1",
			key:       mockKey,
			wantedErr: nil,
		},
		{
			name:      "memory_delete_test2",
			key:       mockAccessKey,
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			store.objects = map[string]*memoryObject{
				mockAccessKey: {data: []byte(mockSecretKey)},
			}
			err := store.DeleteObject(context.TODO(), tt.key)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestMemoryStore_HeadObjectSuccess(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		wantedErr error
	}{
		{
			name:      "memory_head_success_test1",
			key:       mockAccessKey,
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			store.objects = map[string]*memoryObject{
				mockAccessKey: {data: []byte(mockSecretKey)},
			}
			obj, err := store.HeadObject(context.TODO(), tt.key)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, mockAccessKey, obj.Key())
		})
	}
}

func TestMemoryStore_HeadObjectError(t *testing.T) {
	cases := []struct {
		name      string
		key       string
		wantedErr error
	}{
		{
			name:      "memory_head_error_test1",
			key:       emptyString,
			wantedErr: ErrInvalidObjectKey,
		},
		{
			name:      "memory_head_error_test2",
			key:       mockKey,
			wantedErr: os.ErrNotExist,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			obj, err := store.HeadObject(context.TODO(), tt.key)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, nil, obj)
		})
	}
}

func TestMemoryStore_ListObjectsSuccess(t *testing.T) {
	cases := []struct {
		name      string
		prefix    string
		marker    string
		delimiter string
		limit     int64
		wantedErr error
	}{
		{
			name:      "memory_list_success_test1",
			prefix:    mockAccessKey,
			marker:    emptyString,
			delimiter: emptyString,
			limit:     1,
			wantedErr: nil,
		},
		{
			name:      "memory_list_success_test2",
			prefix:    emptyString,
			marker:    "a",
			delimiter: emptyString,
			limit:     1,
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			store.objects = map[string]*memoryObject{
				mockAccessKey: {data: []byte(mockSecretKey)},
				"test":        {data: []byte("hello")},
			}
			objs, err := store.ListObjects(context.TODO(), tt.prefix, tt.marker, tt.delimiter, tt.limit)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, mockAccessKey, objs[0].Key())
		})
	}
}

func TestMemoryStore_ListObjectsError(t *testing.T) {
	cases := []struct {
		name      string
		delimiter string
		wantedErr error
	}{
		{
			name:      "memory_list_error_test1",
			delimiter: mockKey,
			wantedErr: ErrUnsupportedDelimiter,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := setupMemoryTest(t)
			objs, err := store.ListObjects(context.TODO(), emptyString, emptyString, tt.delimiter, 0)
			assert.Equal(t, tt.wantedErr, err)
			assert.Equal(t, 0, len(objs))
		})
	}
}

func TestMemoryStore_ListAllObjects(t *testing.T) {
	store := setupMemoryTest(t)
	_, err := store.ListAllObjects(context.TODO(), emptyString, emptyString)
	assert.Equal(t, ErrUnsupportedMethod, err)
}
