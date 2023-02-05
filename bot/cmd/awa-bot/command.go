package awabot

import (
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tashima42/awa-bot/pkg/db"
	"github.com/tashima42/awa-bot/pkg/telegram"
)

var debug bool

func InitCommand(telegramApiToken string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "awa-bot",
		Short: "Awa bot helps you track your water consumption",
		Long:  "Awa bot helps you track your water consumption",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("running awa-bot command")
			var conf db.Config
			conf.FromEnv()
			log.Println("opening database")
			repo, err := db.Open(conf)
			if err != nil {
				return errors.Wrap(err, "failed to open database")
			}

			log.Print("creating telegram bot")
			telegram, err := telegram.NewBot(debug, telegramApiToken, repo)
			if err != nil {
				return errors.Wrap(err, "failed to start telegram bot")
			}
			log.Print("configuring telegram bot")
			telegram.ConfigBot()

			log.Print("starting to handle updates on telegram bot")
			telegram.HandleUpdates()
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Telegram debug mode")

	return rootCmd
}
