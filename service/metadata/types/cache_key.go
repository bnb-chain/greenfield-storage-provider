package types

var (
	BucketKeyPrefix = []byte{0x11}
	ObjectKeyPrefix = []byte{0x12}
)

func (m *Bucket) GetCacheKey() []byte {
	return append(BucketKeyPrefix, m.BucketInfo.Id.Bytes()...)
}

func (m *Object) GetCacheKey() []byte {
	return append(ObjectKeyPrefix, m.ObjectInfo.Id.Bytes()...)
}
