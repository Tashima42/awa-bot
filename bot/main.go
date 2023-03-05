package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tashima42/awa-bot/bot/api/cmd/awa-api"
	awabot "github.com/tashima42/awa-bot/bot/cmd/awa-bot"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"log"
)

func main() {
	log.Println("running awa-bot command")
	var conf db.Config
	conf.FromEnv()
	log.Println("opening database")
	repo, err := db.Open(conf)
	if err != nil {
		log.Panic(errors.Wrap(err, "failed to open database"))
	}
	if repo == nil {
		log.Panic(errors.Wrap(err, "repo is nil"))
	}
	_, _, err = repo.Up()
	if err != nil && err.Error() != "no change" {
		log.Panic(errors.Wrap(err, "failed to run migrations"))
	}
	rootCmd := &cobra.Command{
		Use:   "awa",
		Short: "awa is a command line tool for awa-bot",
		Long:  "awa is a command line tool for awa-bot",
	}
	log.Println("adding bot command")
	rootCmd.AddCommand(awabot.Command(repo))
	log.Printf("add api command")
	rootCmd.AddCommand(awaapi.Command(repo))
	log.Println("executing command")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to execute command"))
	}
}
