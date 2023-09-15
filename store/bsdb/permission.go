package bsdb

import (
	"errors"
	"time"

	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetPermissionByResourceAndPrincipal get permission by resource type & ID, principal type & value
func (b *BsDBImpl) GetPermissionByResourceAndPrincipal(resourceType, principalType, principalValue string, resourceID common.Hash) (*Permission, error) {
	var (
		permission *Permission
		err        error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	err = b.db.Table((&Permission{}).TableName()).
		Select("*").
		Where("resource_type = ? and resource_id = ? and principal_type = ? and principal_value = ? and removed = false", resourceType, resourceID, permtypes.PrincipalType_value[principalType], principalValue).
		Take(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return permission, err
}

// GetStatementsByPolicyID get statements info by a policy id
func (b *BsDBImpl) GetStatementsByPolicyID(policyIDList []common.Hash, includeRemoved bool) ([]*Statement, error) {
	var (
		statements []*Statement
		err        error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includeRemoved {
		err = b.db.Table((&Statement{}).TableName()).
			Select("*").
			Where("policy_id in ?", policyIDList).
			Find(&statements).Error
	} else {
		err = b.db.Table((&Statement{}).TableName()).
			Select("*").
			Where("policy_id in ? and removed = false", policyIDList).
			Find(&statements).Error
	}
	return statements, err
}

// GetPermissionsByResourceAndPrincipleType get permission by resource type & ID, principal type
func (b *BsDBImpl) GetPermissionsByResourceAndPrincipleType(resourceType, principalType string, resourceID common.Hash, includeRemoved bool) ([]*Permission, error) {
	var (
		permissions []*Permission
		err         error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includeRemoved {
		err = b.db.Table((&Permission{}).TableName()).
			Select("*").
			Where("resource_type = ? and resource_id = ? and principal_type = ?", resourceType, resourceID, permtypes.PrincipalType_value[principalType]).
			Find(&permissions).Error
	} else {
		err = b.db.Table((&Permission{}).TableName()).
			Select("*").
			Where("resource_type = ? and resource_id = ? and principal_type = ? and removed = false", resourceType, resourceID, permtypes.PrincipalType_value[principalType]).
			Find(&permissions).Error
	}
	return permissions, err
}

// Eval is used to evaluate the execution results of permission policies.
func (p Permission) Eval(action permtypes.ActionType, blockTime time.Time, opts *permtypes.VerifyOptions, statements []*Statement) permtypes.Effect {
	var (
		allowed bool
		e       permtypes.Effect
	)

	// 1. the policy is expired, need delete
	if p.ExpirationTime != 0 && time.Unix(p.ExpirationTime, 0).Before(blockTime) {
		// Notice: We do not actively delete policies that expire for users.
		return permtypes.EFFECT_UNSPECIFIED
	}

	// 2. check all the statements
	for _, s := range statements {
		if s.ExpirationTime != 0 && time.Unix(s.ExpirationTime, 0).Before(blockTime) {
			continue
		}
		e = s.Eval(action, opts)
		if e == permtypes.EFFECT_DENY {
			return permtypes.EFFECT_DENY
		} else if e == permtypes.EFFECT_ALLOW {
			allowed = true
		}
	}

	if allowed {
		return permtypes.EFFECT_ALLOW
	}
	return permtypes.EFFECT_UNSPECIFIED
}

// ListObjectPolicies list policies by object info
func (b *BsDBImpl) ListObjectPolicies(objectID common.Hash, actionType permtypes.ActionType, startAfter common.Hash, limit int) ([]*Permission, error) {
	var (
		permissions  []*Permission
		actionValues []int
		now          int64
		err          error
	)

	startTime := time.Now()
	methodName := currentFunction()

	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	now = time.Now().Unix()
	actionValues = PossibleValuesForAction(actionType)
	if len(actionValues) == 0 {
		return nil, nil
	}

	err = b.db.Table((&Permission{}).TableName()).
		Select("*").
		Joins("INNER JOIN statements ON statements.policy_id = permission.policy_id").
		Where(`
			permission.resource_type = ? AND 
			permission.resource_id = ? AND 
			permission.policy_id > ? AND 
			(permission.expiration_time > ? OR permission.expiration_time = 0) AND 
			permission.removed = false AND 
			(statements.expiration_time > ? OR statements.expiration_time = 0)AND 
			statements.action_value in (?) AND 
			statements.effect = "EFFECT_ALLOW" AND
			statements.removed = false`,
			gnfdresource.RESOURCE_TYPE_OBJECT.String(),
			objectID,
			startAfter,
			now,
			now,
			actionValues).
		Order("permission.policy_id").
		Limit(limit).
		Find(&permissions).Error

	return permissions, err
}
