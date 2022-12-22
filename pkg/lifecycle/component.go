package lifecycle

import "context"

// Component provides two methods to implement a component
type Component interface {
	Name() string
	Init(ctx context.Context) error
}

type component struct {
	name string
	fn   func(ctx context.Context) error
}

// NewComponent returns an instance of one component
func NewComponent(name string, fn func(ctx context.Context) error) Component {
	return &component{
		name: name,
		fn:   fn,
	}
}

// Name describes the name of one component
func (c *component) Name() string {
	return c.name
}

// Init initializes one component
func (c *component) Init(ctx context.Context) error {
	return c.fn(ctx)
}
