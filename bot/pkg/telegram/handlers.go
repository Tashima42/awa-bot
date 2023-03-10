package telegram

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tashima42/awa-bot/bot/pkg/auth"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"log"
	"strconv"
	"strings"
	"time"
)

func (t *Telegram) RegisterHandlers() {
	t.AddHandler(t.helpHandler())
	t.AddHandler(t.startHandler())
	t.AddHandler(t.statusHandler())
	t.AddHandler(t.askForRegisterWaterKeyboardHandler())
	t.AddHandler(t.askForRemoveWaterKeyboardHandler())
	t.AddHandler(t.registerWaterHandler())
	t.AddHandler(t.askForCompetitionDurationKeyboardHandler())
	t.AddHandler(t.startCompetitionHandler())
	t.AddHandler(t.competitionHandler())
	t.AddHandler(t.enterCompetitionHandler())
	t.AddHandler(t.askForRegisterGoalKeyboardHandler())
	t.AddHandler(t.registerGoalHandler())
	t.AddHandler(t.goalHandler())
	t.AddHandler(t.registerApiKeyHandler())
	t.AddHandler(t.deleteApiKeyHandler())
	t.AddHandler(t.userIDHandler())
	t.AddHandler(t.apiInstructionsHandler())
}

func (t *Telegram) askForCompetitionDurationKeyboardHandler() (string, Handler) {
	return "new_competition", Handler{
		command:  true,
		hydrate:  false,
		callback: false,
		exec: func(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
			var keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("1 day", "startCompetitionCallback|1"),
					tgbotapi.NewInlineKeyboardButtonData("2 days", "startCompetitionCallback|2"),
					tgbotapi.NewInlineKeyboardButtonData("3 days", "startCompetitionCallback|3"),
					tgbotapi.NewInlineKeyboardButtonData("4 days", "startCompetitionCallback|4"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("5 days", "startCompetitionCallback|5"),
					tgbotapi.NewInlineKeyboardButtonData("6 days", "startCompetitionCallback|6"),
					tgbotapi.NewInlineKeyboardButtonData("1 week", "startCompetitionCallback|7"),
					tgbotapi.NewInlineKeyboardButtonData("2 weeks", "startCompetitionCallback|14"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("3 weeks", "startCompetitionCallback|21"),
					tgbotapi.NewInlineKeyboardButtonData("1 month", "startCompetitionCallback|30"),
					tgbotapi.NewInlineKeyboardButtonData("2 months", "startCompetitionCallback|60"),
					tgbotapi.NewInlineKeyboardButtonData("3 months", "startCompetitionCallback|90"),
				),
			)
			t.SendMessage(message.Chat.ID, "Competition duration", &keyboard)
			return nil
		},
	}
}

func (t *Telegram) startCompetitionHandler() (string, Handler) {
	return "startCompetitionCallback", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, callbackQuery *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tx, err := t.repo.BeginTxx(ctx, &sql.TxOptions{})
			if err != nil {
				return errors.Wrap(err, "failed to start transaction")
			}
			competition, err := t.repo.GetCompetitionByChatTxx(tx, message.Chat.ID)
			if err != nil && err != sql.ErrNoRows {
				return errors.Wrap(err, "failed to get competition")
			}
			if competition != nil {
				t.SendMessage(message.Chat.ID, "There is already a competition running, use /competition to see the status", nil)
				return nil
			}
			days, err := strconv.Atoi(strings.Split(callbackQuery.Data, "|")[1])
			if err != nil {
				return errors.Wrap(err, "failed to start competition")
			}
			log.Printf("starting competition for %d days in chat %d", days, message.Chat.ID)
			err = t.repo.RegisterCompetition(ctx, db.Competition{
				Users:     []string{tgCtx.user.Id},
				ChatID:    message.Chat.ID,
				StartDate: time.Now(),
				EndDate:   time.Now().AddDate(0, 0, days),
			})
			if err != nil {
				return err
			}
			t.SendMessage(message.Chat.ID, "Competition started", nil)
			return nil
		},
	}
}

func (t *Telegram) competitionHandler() (string, Handler) {
	return "competition", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			competition, err := t.repo.GetCompetitionByChat(ctx, message.Chat.ID)
			if competition == nil || err != nil {
				return errors.Wrap(err, "failed to find competition in this chat")
			}
			daysRemaining := competition.EndDate.Sub(competition.StartDate).Hours() / 24
			// TODO: include participants points and ranking
			t.SendMessage(message.Chat.ID, fmt.Sprintf(`Competition:
				Started: %s
				Ends: %s
				Days remaining: %d`,
				competition.StartDate.Format("2006-01-02"),
				competition.EndDate.Format("2006-01-02"),
				int(daysRemaining)),
				nil,
			)
			return nil
		},
	}
}

func (t *Telegram) enterCompetitionHandler() (string, Handler) {
	return "enter_competition", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tx, err := t.repo.BeginTxx(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "failed to begin db transaction")
			}
			competition, err := t.repo.GetCompetitionByChatTxx(tx, message.Chat.ID)
			if competition == nil || err != nil {
				return errors.Wrap(db.Rollback(tx, err), "failed to find competition in this chat")
			}
			if competition.EndDate.Before(time.Now()) {
				return errors.Wrap(db.Rollback(tx, err), "competition has ended")
			}
			if competition.StartDate.After(time.Now()) {
				return errors.Wrap(db.Rollback(tx, err), "competition has not started yet")
			}
			if competition.IsUserRegistered(tgCtx.user.Id) {
				log.Printf("user %s is already registered in competition %s", tgCtx.user.Id, competition.Id)
				t.SendMessage(message.Chat.ID, "You are already registered in this competition", nil)
				return errors.Wrap(db.Rollback(tx, err), "user is already registered in this competition")
			}
			err = t.repo.RegisterUsersInCompetitionTxx(tx, []string{tgCtx.user.Id}, competition.Id)
			if err != nil {
				return errors.Wrap(db.Rollback(tx, err), "failed to register user in competition")
			}
			err = tx.Commit()
			if err != nil {
				return err
			}
			t.SendMessage(message.Chat.ID, "Registered user", nil)
			return nil
		},
	}
}

func (t *Telegram) askForRegisterWaterKeyboardHandler() (string, Handler) {
	return "water", Handler{
		command: true,
		hydrate: true,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			waterStrings := strings.Split(message.Text, " ")
			if len(waterStrings) > 1 {
				amount, err := strconv.Atoi(waterStrings[1])
				if err != nil {
					return err
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				err = t.repo.RegisterWater(ctx, db.Water{UserId: tgCtx.user.Id, Amount: amount})
				if err != nil {
					return err
				}
				msg := fmt.Sprintf("Great, %s, added %dml to your goal", message.From.UserName, amount)
				t.SendMessage(message.Chat.ID, msg, nil)
				return nil
			}
			var keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("100ml", "waterCallback|100"),
					tgbotapi.NewInlineKeyboardButtonData("200ml", "waterCallback|200"),
					tgbotapi.NewInlineKeyboardButtonData("300ml", "waterCallback|300"),
					tgbotapi.NewInlineKeyboardButtonData("400ml", "waterCallback|400"),
					tgbotapi.NewInlineKeyboardButtonData("500ml", "waterCallback|500"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("550ml", "waterCallback|550"),
					tgbotapi.NewInlineKeyboardButtonData("600ml", "waterCallback|600"),
					tgbotapi.NewInlineKeyboardButtonData("700ml", "waterCallback|700"),
					tgbotapi.NewInlineKeyboardButtonData("800ml", "waterCallback|800"),
					tgbotapi.NewInlineKeyboardButtonData("900ml", "waterCallback|900"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("1L", "waterCallback|1000"),
					tgbotapi.NewInlineKeyboardButtonData("2L", "waterCallback|2000"),
					tgbotapi.NewInlineKeyboardButtonData("3L", "waterCallback|3000"),
					tgbotapi.NewInlineKeyboardButtonData("4L", "waterCallback|4000"),
					tgbotapi.NewInlineKeyboardButtonData("5L", "waterCallback|5000"),
				),
			)
			t.SendMessage(message.Chat.ID, "Amount of water", &keyboard)
			return nil
		},
	}
}

func (t *Telegram) askForRemoveWaterKeyboardHandler() (string, Handler) {
	return "remove_water", Handler{
		command: true,
		hydrate: false,
		exec: func(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
			var keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("100ml", "waterCallback|-100"),
					tgbotapi.NewInlineKeyboardButtonData("200ml", "waterCallback|-200"),
					tgbotapi.NewInlineKeyboardButtonData("300ml", "waterCallback|-300"),
					tgbotapi.NewInlineKeyboardButtonData("400ml", "waterCallback|-400"),
					tgbotapi.NewInlineKeyboardButtonData("500ml", "waterCallback|-500"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("550ml", "waterCallback|-550"),
					tgbotapi.NewInlineKeyboardButtonData("600ml", "waterCallback|-600"),
					tgbotapi.NewInlineKeyboardButtonData("700ml", "waterCallback|-700"),
					tgbotapi.NewInlineKeyboardButtonData("800ml", "waterCallback|-800"),
					tgbotapi.NewInlineKeyboardButtonData("900ml", "waterCallback|-900"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("1L", "waterCallback|-1000"),
					tgbotapi.NewInlineKeyboardButtonData("2L", "waterCallback|-2000"),
					tgbotapi.NewInlineKeyboardButtonData("3L", "waterCallback|-3000"),
					tgbotapi.NewInlineKeyboardButtonData("4L", "waterCallback|-4000"),
					tgbotapi.NewInlineKeyboardButtonData("5L", "waterCallback|-5000"),
				),
			)
			t.SendMessage(message.Chat.ID, "How much do you want to remove?", &keyboard)
			return nil
		},
	}
}

func (t *Telegram) askForRegisterGoalKeyboardHandler() (string, Handler) {
	return "new_goal", Handler{
		command: true,
		hydrate: true,
		exec: func(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
			var keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("1L", "newGoalCallback|1000"),
					tgbotapi.NewInlineKeyboardButtonData("1.25L", "newGoalCallback|1250"),
					tgbotapi.NewInlineKeyboardButtonData("1.5L", "newGoalCallback|1500"),
					tgbotapi.NewInlineKeyboardButtonData("1.75L", "newGoalCallback|1750"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("2L", "newGoalCallback|2000"),
					tgbotapi.NewInlineKeyboardButtonData("2.25L", "newGoalCallback|2250"),
					tgbotapi.NewInlineKeyboardButtonData("2.5L", "newGoalCallback|2500"),
					tgbotapi.NewInlineKeyboardButtonData("2.75L", "newGoalCallback|2750"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("3L", "newGoalCallback|3000"),
					tgbotapi.NewInlineKeyboardButtonData("3.25L", "newGoalCallback|3250"),
					tgbotapi.NewInlineKeyboardButtonData("3.5L", "newGoalCallback|3500"),
					tgbotapi.NewInlineKeyboardButtonData("3.75L", "newGoalCallback|3750"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("4L", "newGoalCallback|4000"),
					tgbotapi.NewInlineKeyboardButtonData("4.25L", "newGoalCallback|4250"),
					tgbotapi.NewInlineKeyboardButtonData("4.5L", "newGoalCallback|4500"),
					tgbotapi.NewInlineKeyboardButtonData("4.75L", "newGoalCallback|4750"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("5L", "newGoalCallback|5000"),
					tgbotapi.NewInlineKeyboardButtonData("5.25L", "newGoalCallback|5250"),
					tgbotapi.NewInlineKeyboardButtonData("5.5L", "newGoalCallback|5500"),
					tgbotapi.NewInlineKeyboardButtonData("5.75L", "newGoalCallback|5750"),
				),
			)
			t.SendMessage(message.Chat.ID, "Drinking goal", &keyboard)
			return nil
		},
	}
}

func (t *Telegram) registerGoalHandler() (string, Handler) {
	return "newGoalCallback", Handler{
		command:  false,
		callback: true,
		hydrate:  true,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, callbackQuery *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			goal, err := strconv.Atoi(strings.Split(callbackQuery.Data, "|")[1])
			if err != nil {
				return errors.Wrap(err, "failed to start register new goal")
			}
			err = t.repo.RegisterGoal(ctx, db.Goal{
				UserID: tgCtx.user.Id,
				Goal:   goal,
			})
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("Done, your new goal is %dml", goal)
			callback := tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data)
			_, err = t.bot.Request(callback)
			if err != nil {
				return err
			}
			t.SendMessage(message.Chat.ID, msg, nil)
			return nil
		},
	}
}

func (t *Telegram) goalHandler() (string, Handler) {
	return "goal", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tx, err := t.repo.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
			if err != nil {
				return errors.Wrap(err, "failed to begin db transaction")
			}
			msg, err := t.goalMessage(tx, tgCtx.user.Id)
			if err != nil {
				return errors.Wrap(err, "failed to get goal message")
			}
			err = tx.Commit()
			if err != nil {
				return err
			}

			t.SendMessage(message.Chat.ID, *msg, nil)
			return nil
		},
	}
}

func (t *Telegram) registerWaterHandler() (string, Handler) {
	return "waterCallback", Handler{
		command:  true,
		hydrate:  true,
		callback: true,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, callbackQuery *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			amount, err := strconv.Atoi(strings.Split(callbackQuery.Data, "|")[1])
			if err != nil {
				return errors.Wrap(err, "failed to register water")
			}
			if tgCtx == nil || tgCtx.user == nil {
				return errors.New("context or user missing")
			}
			tx, err := t.repo.BeginTxx(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "failed to begin db transaction")
			}
			err = t.repo.RegisterWaterTxx(tx, db.Water{UserId: tgCtx.user.Id, Amount: amount})
			if err != nil {
				return err
			}
			goalMsg, err := t.goalMessage(tx, tgCtx.user.Id)
			if err != nil {
				return errors.Wrap(err, "failed to get goal message")
			}
			err = tx.Commit()
			if err != nil {
				return err
			}

			msg := fmt.Sprintf("Great, %s, added %dml to your goal", callbackQuery.From.UserName, amount)
			if amount < 0 {
				msg = fmt.Sprintf("Ok, %s, removed %dml from your goal", callbackQuery.From.UserName, amount)
			}
			callback := tgbotapi.NewCallback(callbackQuery.ID, callbackQuery.Data)
			_, err = t.bot.Request(callback)
			if err != nil {
				return err
			}
			t.SendMessage(message.Chat.ID, msg, nil)
			t.SendMessage(message.Chat.ID, *goalMsg, nil)
			return nil
		},
	}
}

func (t *Telegram) registerApiKeyHandler() (string, Handler) {
	return "apikey", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			if message.Chat.Type != "private" {
				return errors.New("this command is only available in private chat")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tx, err := t.repo.BeginTxx(ctx, nil)
			if err != nil {
				return errors.Wrap(err, "failed to begin db transaction")
			}
			if tgCtx == nil || tgCtx.user == nil {
				return errors.New("context or user missing")
			}
			_, err = t.repo.GetApiKeyByUserIdTxx(tx, tgCtx.user.Id)
			if err != nil {
				if !strings.Contains(err.Error(), "no rows in result set") {
					return err
				}
			} else {
				t.SendMessage(message.Chat.ID, "You already have an api key, go back in your history to get it, or use /delete_apikey to invalidate and get a new one", nil)
				return nil
			}
			apiKey := auth.NewUUID()
			apiKeyHash, err := t.hashHelper.Hash(apiKey)
			if err != nil {
				return err
			}
			err = t.repo.RegisterApiKeyTxx(tx, db.Auth{UserID: tgCtx.user.Id, ApiKey: apiKeyHash})
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					return nil
				}
				return err
			}
			err = tx.Commit()
			if err != nil {
				return err
			}
			t.SendMessage(message.Chat.ID, "Done, your api key has been registered", nil)
			t.SendMessage(message.Chat.ID, apiKey, nil)
			t.SendMessage(message.Chat.ID, "Please keep it safe, you will not be able to get it back", nil)
			t.SendMessage(message.Chat.ID, "You can use /delete_apikey to invalidate it and get a new one", nil)
			t.SendMessage(message.Chat.ID, "For api usage instructions, check /api_instructions", nil)
			return nil
		},
	}
}

func (t *Telegram) deleteApiKeyHandler() (string, Handler) {
	return "delete_apikey", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			if message.Chat.Type != "private" {
				return errors.New("this command is only available in private chat")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tgCtx == nil || tgCtx.user == nil {
				return errors.New("context or user missing")
			}
			err := t.repo.DeleteApiKey(ctx, tgCtx.user.Id)
			if err != nil {
				if err == sql.ErrNoRows {
					t.SendMessage(message.Chat.ID, "You don't have an api key, use /apikey to get one", nil)
					return nil
				}
				return err
			}
			t.SendMessage(message.Chat.ID, "Done, your api key was removed", nil)
			return nil
		},
	}
}

func (t *Telegram) helpHandler() (string, Handler) {
	return "help", Handler{
		command: true,
		hydrate: false,
		exec:    t.helpMessage,
	}
}

func (t *Telegram) startHandler() (string, Handler) {
	return "start", Handler{
		command: true,
		hydrate: false,
		exec:    t.helpMessage,
	}
}

func (t *Telegram) helpMessage(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
	t.SendMessage(
		message.Chat.ID,
		`Commands:
/help: see this message

/water: Use this to register how much water you just had in mls
/remove_water: Remove water if you added by accident

/new_competition: Start a new group competition (needs to be in a group chat)
/competition: Get group chat competition
/enter_competition: Enter the competition in this group chat

/new_goal: Set a new drinking goal
/goal: See hou you're doing with your goal today

/apikey: Get your api key to use the api
/delete_apikey: Delete your api key
/userid: Get your internal bot user id
/api_instructions: Get Api usage instructions
/create_auth_code: Create auth code
`,
		nil,
	)
	return nil
}

func (t *Telegram) statusHandler() (string, Handler) {
	return "status", Handler{
		command: true,
		hydrate: false,
		exec: func(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
			t.SendMessage(message.Chat.ID, "(┛ಠ_ಠ)┛彡┻━┻", nil)
			return nil
		},
	}
}

func (t *Telegram) userIDHandler() (string, Handler) {
	return "userid", Handler{
		command: true,
		hydrate: true,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			if tgCtx == nil || tgCtx.user == nil {
				return errors.New("context or user missing")
			}
			t.SendMessage(message.Chat.ID, "Your user id is:", nil)
			t.SendMessage(message.Chat.ID, tgCtx.user.Id, nil)
			return nil
		},
	}
}

func (t *Telegram) apiInstructionsHandler() (string, Handler) {
	return "api_instructions", Handler{
		command: true,
		hydrate: true,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, _ *tgbotapi.CallbackQuery) error {
			if tgCtx == nil || tgCtx.user == nil {
				return errors.New("context or user missing")
			}
			t.SendMessage(message.Chat.ID, `This bot works alongside a GraphQL API, you can check the schema and test it in the playground, at https://awa.tashima.space

To use it, you'll need to first add two cookies in your browser:

apikey: {/apikey}
userid: {/userid}

Instead of using cookies, you can also send both as Headers, using the keys:

Authorization: {/apikey}
X-UserID: {/userid}`, nil)
			return nil
		},
	}
}

func (t *Telegram) goalMessage(tx *sqlx.Tx, userID string) (*string, error) {
	goal, err := t.repo.GetGoalByUserTxx(tx, userID)
	if goal == nil || err != nil {
		return nil, errors.Wrap(err, "failed to find goal for this user, use /new_goal to set a new one")
	}
	amount, err := t.repo.GetUserAmountTxx(tx, userID, db.Today)
	if amount == nil || err != nil {
		return nil, errors.Wrap(err, "failed ot get user amount for today, try getting some water")
	}
	diff := percentageMissing(int(*amount), goal.Goal)
	msg := fmt.Sprintf("Goal: %dml\nDrinked today: %dml\n%s", goal.Goal, *amount, percentageBar(diff))
	return &msg, nil
}

func percentageMissing(current int, goal int) float64 {
	diff := float64(goal - current)
	return 100 - ((100 * diff) / float64(goal))
}

func percentageBar(percentage float64) string {
	bars := int(percentage / 10)
	f := "■"
	e := "□"
	bar := ""
	for i := 0; i < 10; i++ {
		if i <= bars && percentage != 0 {
			bar = bar + f
		} else {
			bar = bar + e
		}
	}
	return fmt.Sprintf("%s %.1f%s", bar, percentage, "%")
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
