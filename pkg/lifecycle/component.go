package lifecycle

import "context"

type Component interface {
	Name() string
	Init(ctx context.Context) error
}

type component struct {
	name string
	fn   func(ctx context.Context) error
}

func NewComponent(name string, fn func(ctx context.Context) error) Component {
	return &component{
		name: name,
		fn:   fn,
	}
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Init(ctx context.Context) error {
	return c.fn(ctx)
}
