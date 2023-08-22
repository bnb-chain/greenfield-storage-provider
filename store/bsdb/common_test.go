package bsdb

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUint32ArrayScan(t *testing.T) {
	var arr Uint32Array
	err := arr.Scan([]byte("1,2,3"))
	assert.NoError(t, err)
	assert.Equal(t, Uint32Array{1, 2, 3}, arr)
}

func TestUint32ArrayValue(t *testing.T) {
	arr := Uint32Array{1, 2, 3}
	val, err := arr.Value()
	assert.NoError(t, err)
	assert.Equal(t, "1,2,3", val)
}

func TestUint32ArrayScanWithNilValue(t *testing.T) {
	var arr Uint32Array
	err := arr.Scan(nil)
	assert.NoError(t, err)
	assert.Nil(t, arr)
}

func TestUint32ArrayScanWithInvalidValue(t *testing.T) {
	var arr Uint32Array
	err := arr.Scan([]byte("1,2,abc")) // abc is not a valid uint32
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan Uint32Array value")
}

func TestUint32ArrayValueWithEmptyArray(t *testing.T) {
	arr := Uint32Array{}
	val, err := arr.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestUint32ArrayScanTypeMismatch(t *testing.T) {
	var array Uint32Array
	invalidInput := 12345 // This is of type int, not []byte

	err := array.Scan(invalidInput)
	expectedErrMsg := fmt.Sprintf("failed to scan Uint32Array value: %v", invalidInput)

	assert.NotNil(t, err, "expected an error due to type mismatch but got nil")
	assert.Equal(t, expectedErrMsg, err.Error(), "unexpected error message")
}

func TestMetadataDatabaseFailureMetrics(t *testing.T) {

	err := errors.New("sample error")
	startTime := time.Now()
	methodName := "testMethod"
	MetadataDatabaseFailureMetrics(err, startTime, methodName)
}

func TestMetadataDatabaseSuccessMetrics(t *testing.T) {
	startTime := time.Now()
	methodName := "testMethod"
	MetadataDatabaseSuccessMetrics(startTime, methodName)
}

func TestCurrentFunction(t *testing.T) {
	funcName := currentFunction()
	assert.Equal(t, "TestCurrentFunction", funcName)
}
