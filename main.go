package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/tibeahx/gpt-helper/openaix"
	"github.com/tibeahx/gpt-helper/proxy"
	"github.com/tibeahx/gpt-helper/telegram"
)

func main() {
	l := logrus.New()
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	ai := openaix.NewOpenAi(os.Getenv(("AI_TOKEN")), l)
	b, err := telegram.NewBot(os.Getenv("BOT_TOKEN"), l, ai)
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		b.Run()
		defer wg.Done()
	}()

	rotation, err := proxy.NewRotation("proxies.json")
	if err != nil {
		log.Fatal(err)
	}
	go rotation.Start(time.Duration(60), &wg)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sig

	l.Info("shutting down...")
	b.Stop()
}
