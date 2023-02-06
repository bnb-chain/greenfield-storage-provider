package redundancy

type RedundancyConfig struct {
	BlockNumber  uint64
	SegmentsSize uint64
	ECcfg        ECConfig
}

type ECConfig struct {
	dataBlocks   int
	parityBlocks int
}

var Redundancy map[int]RedundancyConfig

type Object struct {
	ObjectInfo *ObjectInfo
	ObjectData []byte
	// ObjectData ObjectPayloadReader
}
type ObjectInfo struct {
	ID         uint64
	objectName string
	ObjectSize uint64
	Redundancy RedundancyConfig
}
