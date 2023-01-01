package hash

import "reflect"

// Lengths of hashes and addresses in bytes.
const (
	// LengthHash is the expected length of the hash
	LengthHash = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
)

var (
	hashT    = reflect.TypeOf(Hash{})
	addressT = reflect.TypeOf(Address{})
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [LengthHash]byte
type Address [AddressLength]byte
