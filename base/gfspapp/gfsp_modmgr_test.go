package gfspapp

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/stretchr/testify/assert"
)

func TestRegisterModularFailure1(t *testing.T) {
	t.Log("Failure case description: module name cannot be blank")
	defer ClearRegisterModules()
	assert.Panics(t, func() {
		RegisterModular("", "", nil)
	})
}

func TestRegisterModularFailure2(t *testing.T) {
	t.Log("Failure case description: module repeated")
	approver := mockApprover{t: t}
	RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, approver.new)
	defer ClearRegisterModules()
	assert.Panics(t, func() {
		RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, nil)
	})
}

func TestGetRegisterModuleDescription(t *testing.T) {
	approver := mockApprover{t: t}
	RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, approver.new)
	defer ClearRegisterModules()
	result := GetRegisterModuleDescription()
	assert.Equal(t, "approver             Handles the ask crate bucket/object and replicates piece approval request.\n", result)
}
