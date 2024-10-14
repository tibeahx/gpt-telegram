package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/tibeahx/gpt-helper/openaix"
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		b.Run()
		defer wg.Done()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-sig

	l.Info("shutting down...")
	b.Stop()
}
