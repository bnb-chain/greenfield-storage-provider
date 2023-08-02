package util

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToUint64(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputUint64, err := StringToUint64(tt.originString)
			if tt.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, tt.wantedUint64, outputUint64)
		})
	}
}

func TestStringToInt64(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputInt64, err := StringToInt64(tt.originString)
			if tt.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, tt.wantedInt64, outputInt64)
		})
	}
}

func TestStringToUint32(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputUint32, err := StringToUint32(tt.originString)
			if tt.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, tt.wantedUint32, outputUint32)
		})
	}
}

func TestStringToInt32(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputInt32, err := StringToInt32(tt.originString)
			if tt.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, tt.wantedInt32, outputInt32)
		})
	}
}

func TestStringToBool(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputBool, err := StringToBool(tt.originString)
			if tt.wantedIsErr {
				require.Error(t, err)
			}
			assert.Equal(t, tt.wantedBool, outputBool)
		})
	}
}

func TestBoolToInt(t *testing.T) {
	cases := []struct {
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
			"true bool",
			true,
			1,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputInt := BoolToInt(tt.originBool)
			assert.Equal(t, tt.wantedInt, outputInt)
		})
	}
}

func TestJoinWithComma(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputString := JoinWithComma(tt.originStringSlice)
			assert.Equal(t, tt.wantedString, outputString)
		})
	}
}

func TestSplitByComma(t *testing.T) {
	cases := []struct {
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
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			outputStringSlice := SplitByComma(tt.originString)
			assert.Equal(t, tt.wantedString, JoinWithComma(outputStringSlice))
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

func TestStringToBytesSlice(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		wantedSlice [][]byte
		wantedIsErr bool
	}{
		{
			name:        "right",
			str:         "48656c6c6f20476f7068657221",
			wantedSlice: [][]uint8([][]uint8{[]uint8{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x47, 0x6f, 0x70, 0x68, 0x65, 0x72, 0x21}}),
			wantedIsErr: false,
		},
		{
			name:        "wrong",
			str:         "string",
			wantedSlice: [][]uint8{[]uint8{}},
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			slice, err := StringToBytesSlice(tt.str)
			assert.Equal(t, tt.wantedSlice, slice)
			if tt.wantedIsErr {
				require.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestUint32SliceToString(t *testing.T) {
	str := Uint32SliceToString([]uint32{1, 2, 3})
	assert.Equal(t, "1,2,3", str)
}

func TestStringToUint32Slice(t *testing.T) {
	cases := []struct {
		name        string
		str         string
		wantedSlice []uint32
		wantedIsErr bool
	}{
		{
			name:        "right",
			str:         "1,2,3",
			wantedSlice: []uint32{1, 2, 3},
			wantedIsErr: false,
		},
		{
			name:        "wrong",
			str:         "string",
			wantedSlice: nil,
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			slice, err := StringToUint32Slice(tt.str)
			assert.Equal(t, tt.wantedSlice, slice)
			if tt.wantedIsErr {
				require.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestStringArrayToUint32Slice(t *testing.T) {
	cases := []struct {
		name        string
		arr         pq.StringArray
		wantedSlice []uint32
		wantedIsErr bool
	}{
		{
			name:        "right",
			arr:         pq.StringArray{"123"},
			wantedSlice: []uint32{0x7b},
			wantedIsErr: false,
		},
		{
			name:        "wrong",
			arr:         pq.StringArray{"test"},
			wantedSlice: nil,
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			slice, err := StringArrayToUint32Slice(tt.arr)
			assert.Equal(t, tt.wantedSlice, slice)
			if tt.wantedIsErr {
				require.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
