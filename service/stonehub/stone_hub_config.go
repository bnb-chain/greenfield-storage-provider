package stonehub

type StoneHubConfig struct {
	Address string
}

var DefaultStoneHubConfig = &StoneHubConfig{
	Address: "127.0.0.1:5323",
}
