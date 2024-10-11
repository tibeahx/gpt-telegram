package main

import (
	"log"
	"os"

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

	b.Run()
}
