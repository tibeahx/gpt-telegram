package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tibeahx/gpt-helper/logger"
	"github.com/tibeahx/gpt-helper/openaix"
	"github.com/tibeahx/gpt-helper/proxy"
	"github.com/tibeahx/gpt-helper/telegram"
)

func main() {
	log := logger.GetLogger()

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	ai := openaix.NewOpenAi(os.Getenv(("AI_TOKEN")))
	b, err := telegram.NewBot(os.Getenv("BOT_TOKEN"), ai)
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		b.Run()
		defer wg.Done()
	}()

	rotation, err := proxy.NewRotation("proxy.json")
	if err != nil {
		log.Fatal(err)
	}
	// configure proxy rotation in minutes
	go rotation.Start(time.Duration(time.Minute*15), &wg)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sig

	log.Info("shutting down...")
	b.Stop()
}
