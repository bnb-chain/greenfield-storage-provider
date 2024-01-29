package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateToDescription(t *testing.T) {
	const Unknown TaskState = -1
	cases := []struct {
		name         string
		state        TaskState
		wantedResult string
	}{
		{
			name:         "1",
			state:        TaskState_TASK_STATE_INIT_UNSPECIFIED,
			wantedResult: "object meta is created onto the chain",
		},
		{
			name:         "2",
			state:        Unknown,
			wantedResult: "-1",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := StateToDescription(tt.state)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestCheckAllowUploadStatus(t *testing.T) {
	const Unknown TaskState = -1
	cases := []struct {
		name         string
		state        TaskState
		wantedResult bool
	}{
		{
			name:         "1",
			state:        TaskState_TASK_STATE_INIT_UNSPECIFIED,
			wantedResult: true,
		},
		{
			name:         "2",
			state:        TaskState_TASK_STATE_UPLOAD_OBJECT_DONE,
			wantedResult: false,
		},
		{
			name:         "3",
			state:        TaskState_TASK_STATE_REPLICATE_OBJECT_DOING,
			wantedResult: false,
		},
		{
			name:         "4",
			state:        TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR,
			wantedResult: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckAllowUploadStatus(tt.state)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}
