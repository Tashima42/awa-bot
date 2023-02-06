package awabot

import (
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tashima42/awa-bot/bot/pkg/telegram"
)

var debug bool

func Command(repo *db.Repo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bot",
		Short: "Awa bot helps you track your water consumption",
		Long:  "Awa bot helps you track your water consumption",
		RunE: func(cmd *cobra.Command, args []string) error {

			log.Print("creating telegram bot")
			telegram, err := telegram.NewBot(debug, repo)
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
