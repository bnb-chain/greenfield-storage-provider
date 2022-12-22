package main

import (
	"context"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/bnb-chain/inscription-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func morning(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "morning!")
}

func evening(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "evening!")
}

type Morning struct{}

func (m Morning) Start(ctx context.Context) error {
	http.HandleFunc("/morning", morning)
	http.ListenAndServe("localhost:8080", nil)
	return nil
}

func (m Morning) Stop(ctx context.Context) error {
	log.Info("Stop morning service...")
	return nil
}

func (m Morning) Name() string {
	return "Morning 123"
}

type Evening struct{}

func (e Evening) Start(ctx context.Context) error {
	http.HandleFunc("/evening", evening)
	http.ListenAndServe("localhost:8081", nil)
	return nil
}

func (e Evening) Stop(ctx context.Context) error {
	log.Info("Stop evening service...")
	return nil
}

func (e Evening) Name() string {
	return "Evening 456"
}

func main() {
	ctx := context.Background()
	l := lifecycle.NewService(5 * time.Second)
	var (
		m Morning
		e Evening
	)
	l.Signals(syscall.SIGINT, syscall.SIGTERM).StartServices(ctx, m, e).Wait(ctx)
	//fmt.Println(syscall.Getpid())
	//ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	//s.StartServices(ctx, s1, s2)
	//defer s.Close()
	//s.Wait()
}
