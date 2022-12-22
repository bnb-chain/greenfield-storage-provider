# Service Lifecycle

This package provides useful method to manage the lifecycle of a service or a group of services with convenient initializations of components.

## Interface

```go
type Component interface {
    Name() string
    Init(ctx context.Context) error
}

type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
```

## Feature

- Start and stop a group of services
- Init function with components
- Monitor signals to stop services

## Example

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/bnb-chain/inscription-storage-provider/pkg/lifecycle"
)

func server1(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "server1!")
}

func server2(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "server2!")
}

type Server1 struct{}

func (s Server1) Start(ctx context.Context) error {
	fmt.Println("Server1 start")
	http.HandleFunc("/server1", server1)
	http.ListenAndServe("localhost:8080", nil)
	return nil
}

func (s Server1) Stop(ctx context.Context) error {
	fmt.Println("Stop server1 service...")
	return nil
}

func (s Server1) Init(ctx context.Context) error { return nil }

func (s Server1) Name() string {
	return "Service: server1"
}

type Server2 struct{}

func (s Server2) Start(ctx context.Context) error {
	fmt.Println("Server2 start")
	http.HandleFunc("/server2", server2)
	http.ListenAndServe("localhost:8081", nil)
	return nil
}

func (s Server2) Stop(ctx context.Context) error {
	fmt.Println("Stop server2 service...")
	return nil
}

func (s Server2) Init(ctx context.Context) error { return nil }

func (s Server2) Name() string {
	return "Service: server2"
}

func main() {
	ctx := context.Background()
	l := lifecycle.NewService(5 * time.Second)
	var (
		s1 = Server1{}
		s2 = Server2{}
	)
	l.Signals(syscall.SIGINT, syscall.SIGTERM).Init(ctx, s1, s2).StartServices(ctx, s1, s2).Wait(ctx)
}
```
