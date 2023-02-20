package spdb

import "gorm.io/gorm"

// Provider stands for storage providers, which are updated by monitoring onchain events.
type Provider struct {
	gorm.Model
	NodeId string
}

type P2PNodeDB interface {
	Get(nodeId string) (Provider, error)
	Create(provider *Provider) error
	Delete(nodeId string) error
	FetchAll() ([]Provider, error)
}
