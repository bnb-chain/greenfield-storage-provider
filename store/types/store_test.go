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
