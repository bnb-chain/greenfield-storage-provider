package bsdb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SecondarySpAddresses defines the addresses of secondary_sps
type SecondarySpAddresses []string

// Scan value into bytes, implements sql.Scanner interface
func (s *SecondarySpAddresses) Scan(value interface{}) error {
	bytesValue, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal SecondarySpAddresses value: %v", value)
	}
	if len(bytesValue) == 0 {
		return nil
	}
	return json.Unmarshal(bytesValue, s)
}

// Value returns []string value, implements driver.Valuer interface
func (s SecondarySpAddresses) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Checksums defines the root hash of the pieces which stored in a SP
type Checksums []string

// Scan value into bytes, implements sql.Scanner interface
func (c *Checksums) Scan(value interface{}) error {
	bytesValue, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal CheckSums value: %v", value)
	}
	if len(bytesValue) == 0 {
		return nil
	}
	return json.Unmarshal(bytesValue, c)
}

// Value returns []string value, implements driver.Valuer interface
func (c Checksums) Value() (driver.Value, error) {
	if len(c) == 0 {
		return nil, nil
	}
	return json.Marshal(c)
}

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
