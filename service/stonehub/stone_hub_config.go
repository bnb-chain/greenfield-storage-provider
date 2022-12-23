package stonehub

type StoneHubConfig struct {
	Addr        string
	SegmentSize uint64
	ECM         uint32
	ECK         uint32
	StoneType   []string
}
