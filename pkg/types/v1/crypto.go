package v1

// These functions export type tags for use with internal/jsontypes.

func (*PublicKey) TypeTag() string           { return "gnfd-sp.crypto.PublicKey" }
func (*PublicKey_Ed25519) TypeTag() string   { return "gnfd-sp.crypto.PublicKey_Ed25519" }
func (*PublicKey_Secp256K1) TypeTag() string { return "gnfd-sp.crypto.PublicKey_Secp256K1" }
