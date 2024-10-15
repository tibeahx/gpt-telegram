package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/tibeahx/gpt-helper/config"
	"github.com/tibeahx/gpt-helper/logger"
	"github.com/tibeahx/gpt-helper/openaix"
	"github.com/tibeahx/gpt-helper/proxy"
	"github.com/tibeahx/gpt-helper/telegram"
)

func main() {
	log := logger.GetLogger()

	rotation, err := proxy.NewRotation("proxy.json")
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.LoadConfig(filepath.Join(".", "config.yaml"))

	clientReady := make(chan struct{})

	ai := openaix.NewOpenAi(cfg, rotation.HttpClient())
	b, err := telegram.NewBot(cfg, ai)
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		b.Run()
		defer wg.Done()
	}()

	go func() {
		rotation.Start(ai, cfg, &wg)
		close(clientReady)
	}()

	<-clientReady

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sig

	log.Info("shutting down...")
	b.Stop()
}
