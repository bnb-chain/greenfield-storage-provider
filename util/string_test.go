package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToUint64(t *testing.T) {
	testCases := []struct {
		name         string
		originString string
		wantedIsErr  bool
		wantedUint64 uint64
	}{
		{
			"invalid integer case",
			"aa100aa",
			true,
			0,
		},
		{
			"100 integer case",
			"100",
			false,
			100,
		},
		// math.MaxUint32 = 4294967295
		{
			"uint32 max integer case",
			"4294967295",
			false,
			4294967295,
		},
		{
			"uint32 max + 100 integer case",
			"4294967395",
			false,
			4294967395,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputUint64, err := StringToUint64(testCase.originString)
			if testCase.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, testCase.wantedUint64, outputUint64)
		})
	}
}

func TestStringToInt64(t *testing.T) {
	testCases := []struct {
		name         string
		originString string
		wantedIsErr  bool
		wantedInt64  int64
	}{
		{
			"invalid integer case",
			"aa100aa",
			true,
			0,
		},
		{
			"100 integer case",
			"100",
			false,
			100,
		},
		// math.MaxInt32 = 2147483647
		{
			"uint32 max integer case",
			"2147483647",
			false,
			2147483647,
		},
		{
			"uint32 max + 100 integer case",
			"2147483747",
			false,
			2147483747,
		},
		{
			"-100 integer case",
			"-100",
			false,
			-100,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputInt64, err := StringToInt64(testCase.originString)
			if testCase.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, testCase.wantedInt64, outputInt64)
		})
	}
}

func TestStringToUint32(t *testing.T) {
	testCases := []struct {
		name         string
		originString string
		wantedIsErr  bool
		wantedUint32 uint32
	}{
		{
			"invalid uint32",
			"-100",
			true,
			0,
		},
		// math.MaxUint32 = 4294967295
		{
			"uint32 max + 100 integer overflow case",
			"4294967395",
			true,
			0,
		},
		{
			"100 integer succeed case",
			"100",
			false,
			100,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputUint32, err := StringToUint32(testCase.originString)
			if testCase.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, testCase.wantedUint32, outputUint32)
		})
	}
}

func TestStringToInt32(t *testing.T) {
	testCases := []struct {
		name         string
		originString string
		wantedIsErr  bool
		wantedInt32  int32
	}{
		{
			"invalid int32",
			"aa100aa",
			true,
			0,
		},
		// math.MaxUint32 = 4294967295
		{
			"uint32 max + 100 integer overflow case",
			"4294967395",
			true,
			0,
		},
		{
			"100 integer succeed case",
			"100",
			false,
			100,
		},
		{
			"-100 integer case",
			"-100",
			false,
			-100,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputInt32, err := StringToInt32(testCase.originString)
			if testCase.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, testCase.wantedInt32, outputInt32)
		})
	}
}

func TestStringToBool(t *testing.T) {
	testCases := []struct {
		name         string
		originString string
		wantedIsErr  bool
		wantedBool   bool
	}{
		{
			"invalid bool",
			"xxx",
			true,
			false,
		},
		{
			"false bool",
			"false",
			false,
			false,
		},
		{
			"true bool",
			"true",
			false,
			true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputBool, err := StringToBool(testCase.originString)
			if testCase.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, testCase.wantedBool, outputBool)
		})
	}
}

func TestBoolToInt(t *testing.T) {
	testCases := []struct {
		name       string
		originBool bool
		wantedInt  int
	}{
		{
			"false bool",
			false,
			0,
		},
		{
			"false bool",
			true,
			1,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputInt := BoolToInt(testCase.originBool)
			assert.Equal(t, testCase.wantedInt, outputInt)
		})
	}
}

func TestJoinWithComma(t *testing.T) {
	testCases := []struct {
		name              string
		originStringSlice []string
		wantedString      string
	}{
		{
			"multi slice",
			[]string{"123", "456"},
			"123,456",
		},
		{
			"one slice",
			[]string{"123"},
			"123",
		},
		{
			"empty slice",
			[]string{},
			"",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputString := JoinWithComma(testCase.originStringSlice)
			assert.Equal(t, testCase.wantedString, outputString)
		})
	}
}

func TestSplitByComma(t *testing.T) {
	testCases := []struct {
		name         string
		originString string
		wantedString string
	}{
		{
			"normal case",
			"a,b,c",
			"a,b,c",
		},
		{
			"trim space case",
			" a ,b , c",
			"a,b,c",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			outputStringSlice := SplitByComma(testCase.originString)
			assert.Equal(t, testCase.wantedString, JoinWithComma(outputStringSlice))
		})
	}
}

func TestUintToString(t *testing.T) {
	assert.Equal(t, "100", Uint64ToString(100))
	assert.Equal(t, "100", Uint32ToString(100))
}

func TestBytesSliceToString(t *testing.T) {
	bytesSlice, err := StringToBytesSlice("12345678,87654321")
	require.NoError(t, err)
	assert.Equal(t, "12345678,87654321", BytesSliceToString(bytesSlice))
}
