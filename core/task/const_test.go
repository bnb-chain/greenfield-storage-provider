package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTKey_String(t *testing.T) {
	var a TKey = "test"
	result := a.String()
	assert.Equal(t, "test", result)
}

func TestTaskTypeName(t *testing.T) {
	cases := []struct {
		name         string
		taskType     TType
		wantedResult string
	}{
		{
			name:         "1",
			taskType:     TypeTaskCreateBucketApproval,
			wantedResult: "CreateBucketApprovalTask",
		},
		{
			name:         "2",
			taskType:     TType(-1),
			wantedResult: "UnknownTask",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := TaskTypeName(tt.taskType)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}
