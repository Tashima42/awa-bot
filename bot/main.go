package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	awabot "github.com/tashima42/awa-bot/cmd/awa-bot"
)

func main() {
	log.Println("starting application")
	telegramApiToken := os.Getenv("TELEGRAM_TOKEN")
	log.Println("initiating command")
	rootCmd := awabot.InitCommand(telegramApiToken)
	log.Println("executing command")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to execute command"))
	}
}
