package model

type Server interface {
	Init(configPath string) bool
	Start() bool
	Join() bool
	Stop() bool
	Description() string
}
