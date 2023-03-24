package bsdb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// OutFlows defines the accumulated outflow rates of the stream account
type OutFlows []types.OutFlow

// Scan value into bytes, implements sql.Scanner interface
func (o *OutFlows) Scan(value interface{}) error {
	bytesValue, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal OutFlows value: %v", value)
	}
	if len(bytesValue) == 0 {
		return nil
	}

	return json.Unmarshal(bytesValue, &o)
}

// Value returns []types.OutFlow value, implements driver.Valuer interface
func (o OutFlows) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	return json.Marshal(o)
}
