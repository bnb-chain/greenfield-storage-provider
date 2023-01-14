# Service Lifecycle

This package provides useful method to manage the lifecycle of a service or a group of services with convenient initializations of components.

## Interface

```go
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

## Feature

- Start and stop a group of services
- Monitor signals to stop services
- Graceful shutdown

## Example

```go
package main

import (
	"context"
	"syscall"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"http_server"
	"rpc_server"
)

func main() {
	ctx := context.Background()
	l := lifecycle.NewService(5 * time.Second)
	l.RegisterServices(http_server, rpc_server)
	l.Signals(syscall.SIGINT, syscall.SIGTERM).Init(ctx).StartServices(ctx).Wait(ctx)
}
```
