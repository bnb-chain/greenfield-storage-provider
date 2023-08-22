package bsdb

import (
	"testing"

	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
)

func TestStatement_Eval(t *testing.T) {
	//regexp.MustCompile function will never return nil.
	//Therefore, we don't need to check for reg == nil.
	//If there is an issue with the regular expression pattern, it will panic with an error message
	tests := []struct {
		name      string
		statement *Statement
		action    permtypes.ActionType
		opts      *permtypes.VerifyOptions
		want      permtypes.Effect
	}{
		{
			name:      "nil opts",
			statement: &Statement{},
			action:    permtypes.ACTION_TYPE_ALL,
			opts:      nil,
			want:      permtypes.EFFECT_UNSPECIFIED,
		},
		{
			name:      "empty resource in opts",
			statement: &Statement{Resources: []string{"test_resource"}},
			action:    permtypes.ACTION_TYPE_ALL,
			opts:      &permtypes.VerifyOptions{Resource: ""},
			want:      permtypes.EFFECT_UNSPECIFIED,
		},
		{
			name:      "nil resources in statement",
			statement: &Statement{Resources: nil},
			action:    permtypes.ACTION_TYPE_ALL,
			opts:      &permtypes.VerifyOptions{Resource: "test_resource"},
			want:      permtypes.EFFECT_UNSPECIFIED,
		},
		{
			name: "action matches - deny effect",
			statement: &Statement{
				ActionValue: 1 << int(permtypes.ACTION_UPDATE_BUCKET_INFO),
				Effect:      permtypes.EFFECT_DENY.String(),
				Resources:   []string{"test_resource"},
			},
			action: permtypes.ACTION_UPDATE_BUCKET_INFO,
			opts:   &permtypes.VerifyOptions{Resource: "test_resource"},
			want:   permtypes.EFFECT_DENY,
		},
		{
			name: "action matches - non deny effect",
			statement: &Statement{
				ActionValue: 1 << int(permtypes.ACTION_UPDATE_BUCKET_INFO),
				Effect:      permtypes.EFFECT_ALLOW.String(),
				Resources:   []string{"test_resource"},
			},
			action: permtypes.ACTION_UPDATE_BUCKET_INFO,
			opts:   &permtypes.VerifyOptions{Resource: "test_resource"},
			want:   permtypes.EFFECT_ALLOW,
		},
		{
			name: "action doesn't match",
			statement: &Statement{
				ActionValue: 1 << ActionTypeMap[permtypes.ACTION_UPDATE_BUCKET_INFO],
				Effect:      permtypes.EFFECT_ALLOW.String(),
			},
			action: permtypes.ACTION_DELETE_BUCKET,
			opts:   &permtypes.VerifyOptions{Resource: "test_resource"},
			want:   permtypes.EFFECT_UNSPECIFIED,
		},
		{
			name: "ACTION_TYPE_ALL matches everything",
			statement: &Statement{
				ActionValue: ^0,
				Effect:      permtypes.EFFECT_ALLOW.String(),
				Resources:   []string{"test_resource"},
			},
			action: permtypes.ACTION_DELETE_BUCKET,
			opts:   &permtypes.VerifyOptions{Resource: "test_resource"},
			want:   permtypes.EFFECT_ALLOW,
		},
		{
			name: "action matches - deny effect",
			statement: &Statement{
				Resources:   []string{"test_resource"},
				ActionValue: int(permtypes.ACTION_UPDATE_BUCKET_INFO),
				Effect:      permtypes.EFFECT_ALLOW.String(),
			},
			action: permtypes.ACTION_DELETE_BUCKET,
			opts:   &permtypes.VerifyOptions{Resource: "non_matching_resource"},
			want:   permtypes.EFFECT_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.statement.Eval(tt.action, tt.opts); got != tt.want {
				t.Errorf("Statement.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
