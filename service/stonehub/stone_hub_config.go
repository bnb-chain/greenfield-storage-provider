package stonehub

type MockStoneHubConfig struct {
	Mock   bool
	JobDB  string
	MetaDB string
}

type StoneHubConfig struct {
	StorageProvider string
	Address         string
	MockConfig      *MockStoneHubConfig
}

var DefaultStoneHubConfig = &StoneHubConfig{
	StorageProvider: "bnb-sp",
	Address:         "127.0.0.1:5323",
}
