package redundancy

type RedundancyConfig struct {
	BlockNumber  uint64
	SegmentsSize uint64
	ECcfg        ECConfig
}

type ECConfig struct {
	DataBlocks   uint
	ParityBlocks uint
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

func initConfig() ECConfig {
	return ECConfig{
		DataBlocks:   4,
		ParityBlocks: 2,
	}
}

func SpiltSegments(object *Object) ([]*Segment, error) {
	return nil, nil
}

func MergeSegments(segments []*Segment, offset, size uint) ([]byte, error) {
	return nil, nil
}
