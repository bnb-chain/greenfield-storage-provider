package bsdb

import (
	"regexp"

	permtypes "github.com/bnb-chain/greenfield/x/permission/types"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
)

// Eval is used to evaluate the execution results of statement policies.
func (s *Statement) Eval(action permtypes.ActionType, opts *permtypes.VerifyOptions) permtypes.Effect {
	// If 'resource' is not nil, it implies that the user intends to access a sub-resource, which would
	// be specified in 's.Resources'. Therefore, if the sub-resource in the statement is nil, we will ignore this statement.
	if opts != nil && opts.Resource != "" && s.Resources == nil {
		return permtypes.EFFECT_UNSPECIFIED
	}
	// If 'resource' is not nil, and 's.Resource' is also not nil, it indicates that we should verify whether
	// the resource that the user intends to access matches any items in 's.Resource'
	if opts != nil && opts.Resource != "" && s.Resources != nil {
		isMatch := false
		for _, res := range s.Resources {
			reg := regexp.MustCompile(res)
			if reg == nil {
				continue
			}
			matchRes := reg.MatchString(opts.Resource)
			if matchRes {
				isMatch = matchRes
				break
			}
		}
		if !isMatch {
			return permtypes.EFFECT_UNSPECIFIED
		}
	}

	actions := make([]permtypes.ActionType, 0)
	for _, v := range metadata.ActionTypeMap {
		if s.ActionValue&(1<<v) == 1<<v {
			actions = append(actions, permtypes.ActionType(v))
		}
	}

	for _, act := range actions {
		if act == action || act == permtypes.ACTION_TYPE_ALL {
			// Action matched, if effect is deny, then return deny
			if s.Effect == permtypes.EFFECT_DENY.String() {
				return permtypes.EFFECT_DENY
			}
			return permtypes.Effect(permtypes.Effect_value[s.Effect])
		}
	}

	return permtypes.EFFECT_UNSPECIFIED
}
