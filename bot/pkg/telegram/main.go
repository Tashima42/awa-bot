package telegram

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/tashima42/awa-bot/bot/pkg/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	command  bool
	hydrate  bool
	callback bool
	exec     func(message *tgbotapi.Message, tgCtx *TgContext, callbackQuery *tgbotapi.CallbackQuery) error
}

type TgContext struct {
	user *db.User
}

type Telegram struct {
	bot          *tgbotapi.BotAPI
	updateConfig tgbotapi.UpdateConfig
	handlers     map[string]Handler
	repo         *db.Repo
}

func NewBot(debug bool, repo *db.Repo) (*Telegram, error) {
	telegramApiToken := os.Getenv("TELEGRAM_TOKEN")
	bot, err := tgbotapi.NewBotAPI(telegramApiToken)
	if err != nil {
		return nil, err
	}
	bot.Debug = debug
	t := Telegram{}
	t.bot = bot
	t.repo = repo
	t.handlers = map[string]Handler{}
	return &t, err
}

func (t *Telegram) ConfigBot() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	t.RegisterHandlers()
}

func (t *Telegram) AddHandler(matcher string, handler Handler) {
	t.handlers[matcher] = handler
}

func (t *Telegram) HandleUpdates() {
	log.Println("starting to get updates")
	updates := t.bot.GetUpdatesChan(t.updateConfig)
	for update := range updates {
		log.Printf("update: %+v", update)
		if update.Message != nil {
			if update.Message.IsCommand() {
				log.Printf("message is command: %s", update.Message.Command())
				handler, ok := t.handlers[update.Message.Command()]
				if !ok {
					log.Println("command handler not found")
					t.SendMessage(update.Message.Chat.ID, "Command not found, try /help", nil)
					continue
				}
				tgCtx := TgContext{}
				if handler.hydrate {
					log.Printf("hydrating user, message from id: %d", update.Message.From.ID)
					err := t.hydrateUser(&tgCtx, update.Message.From.ID, update.Message.From.UserName)
					if err != nil {
						log.Println(errors.Wrapf(err, "failed to hydrate user, chat id: %d", update.Message.Chat.ID))
						t.SendMessage(update.Message.Chat.ID, err.Error(), nil)
						continue
					}
				}
				log.Println("exec handler")
				err := handler.exec(update.Message, &tgCtx, nil)
				if err != nil {
					log.Println(errors.Wrapf(err, "failed while executing handler"))
					t.SendMessage(update.Message.Chat.ID, err.Error(), nil)
					continue
				}
			}
		}
		if update.CallbackQuery != nil {
			log.Printf("CallbackQuery: %+v", update.CallbackQuery)
			log.Printf("Callback Data: %s", update.CallbackQuery.Data)
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := t.bot.Request(callback); err != nil {
				log.Println(errors.Wrap(err, "failed to handle callback"))
				t.SendMessage(update.Message.Chat.ID, "failed to handle callback", nil)
				continue
			}
			log.Println("getting matcher")
			matcher := strings.Split(update.CallbackQuery.Data, "|")[0]
			log.Printf("matcher: %s", matcher)
			handler, ok := t.handlers[matcher]
			if !ok {
				log.Println("handler not found")
				t.SendMessage(update.Message.Chat.ID, fmt.Sprintf("couldn't find '%s' callback handler", matcher), nil)
				continue
			}
			tgCtx := TgContext{}
			if handler.hydrate {
				log.Println("hydrating")
				err := t.hydrateUser(&tgCtx, update.CallbackQuery.From.ID, update.CallbackQuery.From.UserName)
				if err != nil {
					log.Println(errors.Wrap(err, "failed to hydrate"))
					t.SendMessage(update.CallbackQuery.Message.Chat.ID, err.Error(), nil)
					continue
				}
			}
			log.Println("executing handler")
			err := handler.exec(update.CallbackQuery.Message, &tgCtx, update.CallbackQuery)
			if err != nil {
				log.Println(errors.Wrap(err, "failed to exec handler"))
				t.SendMessage(update.CallbackQuery.Message.Chat.ID, err.Error(), nil)
				continue
			}
		}
	}
}

func (t *Telegram) SendMessage(chatID int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	if _, err := t.bot.Send(msg); err != nil {
		fmt.Println(err)
	}
}

func (t *Telegram) hydrateUser(tgCtx *TgContext, telegramID int64, name string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := t.repo.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	user, err := t.repo.GetUserTxx(tx, telegramID)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(db.Rollback(tx, err), "failed to get user")
	}
	if user != nil && err != sql.ErrNoRows && user.TelegramID != 0 {
		tgCtx.user = user
		return tx.Commit()
	}
	err = t.repo.RegisterUserTxx(tx, db.User{
		TelegramID: telegramID,
		Name:       name,
	})
	if err != nil {
		fmt.Println(err)
		return errors.Wrap(db.Rollback(tx, err), "failed to create user")
	}
	user, err = t.repo.GetUserTxx(tx, telegramID)
	if err != nil {
		fmt.Println(err)
		return errors.Wrap(db.Rollback(tx, err), "failed to get user")
	}
	if user != nil && user.TelegramID != 0 {
		tgCtx.user = user
	}
	return tx.Commit()
}
