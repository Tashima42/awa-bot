package telegram

import (
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tashima42/awa-bot/bot/pkg/db"
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
}

func (t *Telegram) askForCompetitionDurationKeyboardHandler() (string, Handler) {
	return "new_competition", Handler{
		command:  true,
		hydrate:  false,
		callback: false,
		exec: func(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
			var keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("1 day", "newCompetitionCallback|1"),
					tgbotapi.NewInlineKeyboardButtonData("2 days", "newCompetitionCallback|2"),
					tgbotapi.NewInlineKeyboardButtonData("3 days", "newCompetitionCallback|3"),
					tgbotapi.NewInlineKeyboardButtonData("4 days", "newCompetitionCallback|4"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("5 days", "newCompetitionCallback|5"),
					tgbotapi.NewInlineKeyboardButtonData("6 days", "newCompetitionCallback|6"),
					tgbotapi.NewInlineKeyboardButtonData("1 week", "newCompetitionCallback|7"),
					tgbotapi.NewInlineKeyboardButtonData("2 weeks", "newCompetitionCallback|14"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("3 weeks", "newCompetitionCallback|21"),
					tgbotapi.NewInlineKeyboardButtonData("1 month", "newCompetitionCallback|30"),
					tgbotapi.NewInlineKeyboardButtonData("2 months", "newCompetitionCallback|60"),
					tgbotapi.NewInlineKeyboardButtonData("3 months", "newCompetitionCallback|90"),
				),
			)
			t.SendMessage(message.Chat.ID, "Competition duration", &keyboard)
			return nil
		},
	}
}

func (t *Telegram) startCompetitionHandler() (string, Handler) {
	return "newCompetitionCallback", Handler{
		command:  true,
		hydrate:  true,
		callback: false,
		exec: func(message *tgbotapi.Message, tgCtx *TgContext, callbackQuery *tgbotapi.CallbackQuery) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			days, err := strconv.Atoi(strings.Split(callbackQuery.Data, "|")[1])
			if err != nil {
				return errors.Wrap(err, "failed to start competition")
			}
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
			err = t.repo.RegisterUserInCompetitionTxx(tx, tgCtx.user.Id, competition.Id)
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
		hydrate: false,
		exec: func(message *tgbotapi.Message, _ *TgContext, _ *tgbotapi.CallbackQuery) error {
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
			t.SendMessage(message.Chat.ID, fmt.Sprintf("Done, your new goal is %dml", goal), nil)
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
			tx, err := t.repo.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
			if err != nil {
				return errors.Wrap(err, "failed to begin db transaction")
			}
			err = t.repo.RegisterWater(ctx, db.Water{UserId: tgCtx.user.Id, Amount: amount})
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

			msg := fmt.Sprintf("Great, %s, added %dml to your goal", callbackQuery.From.FirstName, amount)
			if amount < 0 {
				msg = fmt.Sprintf("Ok, %s, removed %dml from your goal", callbackQuery.From.FirstName, amount)
			}
			t.SendMessage(message.Chat.ID, msg, nil)
			t.SendMessage(message.Chat.ID, *goalMsg, nil)
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
/goal: See hou you're doing with your goal today'`,
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
