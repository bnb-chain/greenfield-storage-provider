package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_rdm(t *testing.T) {
	cases := []struct {
		name         string
		n            int
		data         string
		wantedLength int
	}{
		{
			"The length of output string is 5",
			5,
			"test",
			5,
		},
		{
			"The length of output string is 0",
			0,
			"test",
			0,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			str := rdm(tt.n, tt.data)
			assert.Equal(t, tt.wantedLength, len(str))
		})
	}
}

func TestRandomNum(t *testing.T) {
	cases := []struct {
		name     string
		n, scope int
	}{
		{
			"Random number is not empty",
			5,
			0,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			num := RandomNum(tt.n, tt.scope)
			assert.GreaterOrEqual(t, num, 0)
		})
	}
}

func TestRandInt64(t *testing.T) {
	cases := []struct {
		name     string
		min, max int64
	}{
		{
			"1",
			1,
			20,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			num := RandInt64(tt.min, tt.max)
			assert.NotEmpty(t, num)
			assert.GreaterOrEqual(t, num, tt.min)
		})
	}
}

func TestRandomString(t *testing.T) {
	cases := []struct {
		name string
		n    int
	}{
		{
			"1",
			5,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			str := RandomString(tt.n)
			assert.NotEmpty(t, str)
			assert.Equal(t, tt.n, len(str))
		})
	}
}

func TestRandomStringToLower(t *testing.T) {
	cases := []struct {
		name string
		n    int
	}{
		{
			"1",
			6,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			str := RandomStringToLower(tt.n)
			assert.NotEmpty(t, str)
			assert.Equal(t, tt.n, len(str))
		})
	}
}

func TestRandomStringToUpper(t *testing.T) {
	cases := []struct {
		name string
		n    int
	}{
		{
			"1",
			7,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			str := RandomStringToUpper(tt.n)
			assert.NotEmpty(t, str)
			assert.Equal(t, tt.n, len(str))
		})
	}
}

func TestRandHexKey(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			"1",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hexKey := RandHexKey()
			assert.NotEmpty(t, hexKey)
		})
	}
}

func Test_randString(t *testing.T) {
	cases := []struct {
		name string
		n    int
	}{
		{
			"1",
			5,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hexKey := randString(tt.n)
			assert.NotEmpty(t, hexKey)
		})
	}
}

func TestGetRandomObjectName(t *testing.T) {
	cases := []struct {
		name         string
		wantedStrLen int
	}{
		{
			"1",
			10,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hexKey := GetRandomObjectName()
			assert.NotEmpty(t, hexKey)
			assert.Equal(t, tt.wantedStrLen, len(hexKey))
		})
	}
}

func TestGetRandomBucketName(t *testing.T) {
	cases := []struct {
		name         string
		wantedStrLen int
	}{
		{
			"1",
			5,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hexKey := GetRandomBucketName()
			assert.NotEmpty(t, hexKey)
			assert.Equal(t, tt.wantedStrLen, len(hexKey))
		})
	}
}

func TestGetRandomGroupName(t *testing.T) {
	cases := []struct {
		name         string
		wantedStrLen int
	}{
		{
			"1",
			7,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hexKey := GetRandomGroupName()
			assert.NotEmpty(t, hexKey)
			assert.Equal(t, tt.wantedStrLen, len(hexKey))
		})
	}
}
