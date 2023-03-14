package metadata

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// SecondarySpAddresses defines the addresses of secondary_sps
type SecondarySpAddresses []string

// Scan value into bytes, implements sql.Scanner interface
func (s *SecondarySpAddresses) Scan(value interface{}) error {
	bytesValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal SecondarySpAddresses value:", value))
	}
	if len(bytesValue) == 0 {
		return nil
	}
	return json.Unmarshal(bytesValue, s)
}

// Value return []string value, implement driver.Valuer interface
func (s SecondarySpAddresses) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

type CheckSums []string

// Scan value into bytes, implements sql.Scanner interface
func (s *CheckSums) Scan(value interface{}) error {
	bytesValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal SecondarySpAddresses value:", value))
	}
	if len(bytesValue) == 0 {
		return nil
	}
	return json.Unmarshal(bytesValue, s)
}

// Value return []string value, implement driver.Valuer interface
func (s CheckSums) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}
