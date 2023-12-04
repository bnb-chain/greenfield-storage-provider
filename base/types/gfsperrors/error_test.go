package gfsperrors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGfSpError_Error(t *testing.T) {
	m := &GfSpError{}
	result := m.Error()
	assert.Equal(t, "{}", result)
}

func TestGfSpError_SetError(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{
			name: "1",
			err:  nil,
		},
		{
			name: "2",
			err:  errors.New("mock error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m := &GfSpError{}
			m.SetError(tt.err)
		})
	}
}

func TestMakeGfSpError(t *testing.T) {
	cases := []struct {
		name         string
		err          error
		wantedResult *GfSpError
	}{
		{
			name:         "1",
			err:          nil,
			wantedResult: nil,
		},
		{
			name:         "2",
			err:          &GfSpError{InnerCode: 0},
			wantedResult: nil,
		},
		{
			name:         "3",
			err:          &GfSpError{InnerCode: 1},
			wantedResult: &GfSpError{InnerCode: 1},
		},
		{
			name: "4",
			err:  errors.New("mock error"),
			wantedResult: &GfSpError{
				CodeSpace:      DefaultCodeSpace,
				HttpStatusCode: int32(http.StatusInternalServerError),
				InnerCode:      int32(DefaultInnerCode),
				Description:    errors.New("mock error").Error(),
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeGfSpError(tt.err)
			fmt.Println(result)
		})
	}
}

func TestRegister(t *testing.T) {
	result := Register("test", http.StatusBadRequest, 0, "mock")
	assert.Equal(t, &GfSpError{CodeSpace: "test", HttpStatusCode: 400, InnerCode: 0, Description: "mock"}, result)
}

func TestGfSpErrorList(t *testing.T) {
	_ = Register("test", http.StatusBadRequest, 1, "mock")
	_ = Register("test", http.StatusInternalServerError, 2, "mock")
	errs := GfSpErrorList()
	assert.NotNil(t, errs)
}
